import { describe, expect, it } from 'vitest'
import { connectedTo } from './graph-path'
import type { Edge } from './graph'
import type { IssueId } from './issue'

const id = (value: string) => value as IssueId

const edge = (from: string, to: string, kind: Edge['kind'] = 'blocks'): Edge => ({
	from: id(from),
	to: id(to),
	kind,
})

const setOf = (edges: readonly Edge[], focus: string) => [...connectedTo(edges, id(focus))].sort()

describe('connectedTo', () => {
	it('includes the focused node itself', () => {
		expect(setOf([], 'a')).toEqual(['a'])
	})

	it('follows blockers upstream', () => {
		expect(setOf([edge('a', 'b')], 'b')).toEqual(['a', 'b'])
	})

	it('follows what the focus blocks downstream', () => {
		expect(setOf([edge('a', 'b')], 'a')).toEqual(['a', 'b'])
	})

	it('walks a chain in both directions', () => {
		const edges = [edge('a', 'b'), edge('b', 'c')]

		expect(setOf(edges, 'b')).toEqual(['a', 'b', 'c'])
	})

	it('reaches the far end of a long chain', () => {
		const edges = [edge('a', 'b'), edge('b', 'c'), edge('c', 'd')]

		expect(setOf(edges, 'd')).toEqual(['a', 'b', 'c', 'd'])
	})

	// Two children of one map are not related just by sharing it. Walking a
	// parent edge downwards would light the whole map, which is what the
	// highlight exists to dim.
	it('leaves a sibling out', () => {
		const edges = [edge('map', 'a', 'parent'), edge('map', 'b', 'parent')]

		expect(setOf(edges, 'a')).toEqual(['a', 'map'])
	})

	// A blocker's other dependents do matter: they are what else is waiting on
	// the same thing.
	it('keeps another issue waiting on the same blocker', () => {
		const edges = [edge('a', 'b'), edge('a', 'c')]

		expect(setOf(edges, 'b')).toEqual(['a', 'b', 'c'])
	})

	it('includes a parent, since it places the node', () => {
		expect(setOf([edge('map', 'a', 'parent')], 'a')).toEqual(['a', 'map'])
	})

	// A relation says two things are worth reading together, which is exactly
	// what a highlight is for.
	it('includes a related node', () => {
		expect(setOf([edge('a', 'b', 'relates')], 'a')).toEqual(['a', 'b'])
	})

	it('does not loop forever on a cycle', () => {
		const edges = [edge('a', 'b'), edge('b', 'a')]

		expect(setOf(edges, 'a')).toEqual(['a', 'b'])
	})
})
