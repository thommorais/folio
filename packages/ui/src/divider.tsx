import { tv, type VariantProps } from '@thom/libs/tv'

const dividerClasses = tv({
	base: 'w-full border-t',
	variants: {
		soft: {
			true: 'border-border/50',
			false: 'border-border',
		},
	},
	defaultVariants: { soft: false },
})

type DividerProps = React.ComponentPropsWithRef<'hr'> & VariantProps<typeof dividerClasses>

export const Divider = ({ soft, className, ...props }: DividerProps) => {
	return <hr {...props} className={dividerClasses({ soft, class: className })} data-id='thom-ui' />
}
