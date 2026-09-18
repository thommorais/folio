package pb

import (
	"context"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
	"folio/folio-core/ports"
)

type TagRepository struct {
	app core.App
}

func NewTagRepository(app core.App) *TagRepository {
	return &TagRepository{app: app}
}

var _ ports.TagRepository = (*TagRepository)(nil)

func toTag(rec *core.Record) domain.Tag {
	return domain.Tag{
		ID:       domain.TagID(rec.Id),
		DomainID: domain.DomainID(rec.GetString("domain")),
		Slug:     rec.GetString("slug"),
		Name:     rec.GetString("name"),
	}
}

func (r *TagRepository) ListByDomain(ctx context.Context, domainID domain.DomainID) ([]domain.Tag, error) {
	records, err := r.app.FindAllRecords(ColTags, dbx.HashExp{"domain": string(domainID)})
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]domain.Tag, 0, len(records))
	for _, rec := range records {
		out = append(out, toTag(rec))
	}
	return out, nil
}

func (r *TagRepository) Ensure(ctx context.Context, domainID domain.DomainID, names []string) ([]domain.Tag, error) {
	if len(names) == 0 {
		return nil, nil
	}
	collection, err := r.app.FindCollectionByNameOrId(ColTags)
	if err != nil {
		return nil, mapErr(err)
	}

	existing, err := r.ListByDomain(ctx, domainID)
	if err != nil {
		return nil, err
	}
	bySlug := make(map[string]domain.Tag, len(existing))
	for _, t := range existing {
		bySlug[t.Slug] = t
	}

	out := make([]domain.Tag, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		slug := rules.Slugify(name)
		if slug == "" || seen[slug] {
			continue
		}
		seen[slug] = true

		if t, ok := bySlug[slug]; ok {
			out = append(out, t)
			continue
		}
		rec := core.NewRecord(collection)
		rec.Set("domain", string(domainID))
		rec.Set("slug", slug)
		rec.Set("name", name)
		if err := r.app.Save(rec); err != nil {
			return nil, mapErr(err)
		}
		tag := toTag(rec)
		bySlug[slug] = tag
		out = append(out, tag)
	}
	return out, nil
}

func (r *TagRepository) TagsOf(ctx context.Context, target ports.TagTarget, ids []string) (map[string][]string, error) {
	if len(ids) == 0 {
		return map[string][]string{}, nil
	}
	join, field := tagJoin(target)

	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}
	links, err := r.app.FindAllRecords(join, dbx.In(field, values...))
	if err != nil {
		return nil, mapErr(err)
	}
	if len(links) == 0 {
		return map[string][]string{}, nil
	}

	tagIDs := make([]any, 0, len(links))
	for _, l := range links {
		tagIDs = append(tagIDs, l.GetString("tag"))
	}
	tags, err := r.app.FindAllRecords(ColTags, dbx.In("id", tagIDs...))
	if err != nil {
		return nil, mapErr(err)
	}
	name := make(map[string]string, len(tags))
	for _, t := range tags {
		name[t.Id] = t.GetString("name")
	}

	out := make(map[string][]string, len(ids))
	for _, l := range links {
		if n, ok := name[l.GetString("tag")]; ok {
			target := l.GetString(field)
			out[target] = append(out[target], n)
		}
	}
	return out, nil
}

func (r *TagRepository) SetTags(ctx context.Context, target ports.TagTarget, id string, domainID domain.DomainID, names []string) error {
	join, field := tagJoin(target)

	tags, err := r.Ensure(ctx, domainID, names)
	if err != nil {
		return err
	}
	want := make(map[string]bool, len(tags))
	for _, t := range tags {
		want[string(t.ID)] = true
	}

	links, err := r.app.FindAllRecords(join, dbx.HashExp{field: id})
	if err != nil {
		return mapErr(err)
	}
	for _, l := range links {
		tagID := l.GetString("tag")
		if want[tagID] {
			delete(want, tagID)
			continue
		}
		if err := r.app.Delete(l); err != nil {
			return mapErr(err)
		}
	}

	if len(want) == 0 {
		return nil
	}
	collection, err := r.app.FindCollectionByNameOrId(join)
	if err != nil {
		return mapErr(err)
	}
	for tagID := range want {
		rec := core.NewRecord(collection)
		rec.Set(field, id)
		rec.Set("tag", tagID)
		if err := r.app.Save(rec); err != nil {
			return mapErr(err)
		}
	}
	return nil
}

func (r *TagRepository) IDsWithTags(ctx context.Context, target ports.TagTarget, domainID domain.DomainID, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	join, field := tagJoin(target)

	slugs := make([]any, 0, len(names))
	for _, n := range names {
		slugs = append(slugs, rules.Slugify(n))
	}
	tags, err := r.app.FindAllRecords(ColTags, dbx.HashExp{"domain": string(domainID)}, dbx.In("slug", slugs...))
	if err != nil {
		return nil, mapErr(err)
	}
	if len(tags) < len(slugs) {
		return []string{}, nil
	}

	counts := make(map[string]int)
	for _, t := range tags {
		links, err := r.app.FindAllRecords(join, dbx.HashExp{"tag": t.Id})
		if err != nil {
			return nil, mapErr(err)
		}
		for _, l := range links {
			counts[l.GetString(field)]++
		}
	}

	out := make([]string, 0, len(counts))
	for id, n := range counts {
		if n == len(tags) {
			out = append(out, id)
		}
	}
	return out, nil
}

func tagJoin(target ports.TagTarget) (string, string) {
	if target == ports.TagEntry {
		return ColEntryTags, "entry"
	}
	return ColIssueTags, "issue"
}
