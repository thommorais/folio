import type { Term } from '_/core/domain/interview'

export const Terms = ({ terms }: { readonly terms: readonly Term[] }) => (
	<section className='mx-6 space-y-2'>
		<h2 className='text-sm font-medium'>Terms</h2>

		{terms.length === 0 ? (
			<p className='text-dim text-sm'>No terms yet.</p>
		) : (
			<dl className='border-border divide-border divide-y border'>
				{terms.map(term => (
					<div key={term.term} className='px-4 py-3'>
						<dt className='text-sm'>{term.term}</dt>
						<dd className='text-dim text-xs'>{term.def}</dd>
						{term.avoid.length > 0 && <dd className='text-dimmer text-xs'>Avoid: {term.avoid.join(', ')}</dd>}
					</div>
				))}
			</dl>
		)}
	</section>
)
