'use client'
import { type ComponentProps, useId } from 'react'
import { tv } from '@thom/libs/tv'
import { useDisabled, useProvidedId, useProvidedLabel } from './fieldset.context'

const dateTypes = ['date', 'datetime-local', 'month', 'time', 'week']
type DateType = (typeof dateTypes)[number]

const inputClasses = tv({
	base: [
		'border-border flex h-9 w-full border bg-transparent px-3 py-1',
		'text-foreground placeholder:text-muted-foreground text-sm',
		'transition-colors',
		'focus-visible:border-ring focus-visible:outline-none',
		'disabled:cursor-not-allowed disabled:opacity-50',
		'aria-invalid:border-destructive',
		'file:border-0 file:bg-transparent file:text-sm file:font-medium',
		'[&:-webkit-autofill]:!bg-transparent [&:-webkit-autofill]:!bg-none [&:-webkit-autofill]:!shadow-none',
	],
	variants: {
		isDate: {
			true: [
				'[&::-webkit-datetime-edit-fields-wrapper]:p-0',
				'[&::-webkit-date-and-time-value]:min-h-[1.5em]',
				'[&::-webkit-datetime-edit]:inline-flex',
				'[&::-webkit-datetime-edit]:p-0',
			],
		},
	},
})

type InputProps = ComponentProps<'input'> & {
	type: 'email' | 'number' | 'password' | 'search' | 'tel' | 'text' | 'url' | DateType
}

export const Input = ({ className, ...props }: InputProps) => {
	const providedDisabled = useDisabled()
	const internalId = useId()
	const providedId = useProvidedId()
	const providedLabel = useProvidedLabel()
	const disabled = providedDisabled ?? props.disabled
	const id = providedId ?? props.id ?? internalId
	const name = providedLabel ?? props.name ?? id

	return (
		<input
			{...props}
			name={name}
			id={id}
			disabled={disabled}
			data-slot='control'
			data-disabled={disabled || undefined}
			data-invalid={props['aria-invalid']}
			className={inputClasses({ isDate: dateTypes.includes(props.type), class: className })}
		/>
	)
}
