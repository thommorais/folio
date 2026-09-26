import { describe, expect, it } from 'vitest'
import { TAG_TONES, tagTone } from './tag-tone'

describe('tagTone', () => {
	it('gives a tag the same tone every time', () => {
		expect(tagTone('frontend')).toBe(tagTone('frontend'))
	})

	it('ignores case and surrounding space', () => {
		expect(tagTone(' Bug ')).toBe(tagTone('bug'))
	})

	it('picks from the palette', () => {
		for (const tag of ['api', 'bug', 'x', '', 'a-very-long-tag-name']) {
			expect(TAG_TONES).toContain(tagTone(tag))
		}
	})

	it('spreads the vocabulary over more than one tone', () => {
		const tones = new Set(['api', 'backend', 'cli', 'db', 'design', 'docs', 'bug', 'perf'].map(tagTone))

		expect(tones.size).toBeGreaterThan(3)
	})
})
