import { Monitor, Moon, Sun } from 'lucide-react'
import { cn } from '@thom/libs/cn'
import type { Theme } from '_/core/ports/theme'
import { useTheme } from '_/app/use-theme'

const options = [
	{ value: 'light', label: 'Light', Icon: Sun },
	{ value: 'dark', label: 'Dark', Icon: Moon },
	{ value: 'system', label: 'System', Icon: Monitor },
] as const satisfies readonly { value: Theme; label: string; Icon: typeof Sun }[]

export const ThemeSwitch = () => {
	const { theme, setTheme } = useTheme()

	return (
		<div className='border-border flex items-center border'>
			{options.map(({ value, label, Icon }) => (
				<button
					key={value}
					type='button'
					aria-label={label}
					aria-pressed={theme === value}
					onClick={() => setTheme(value)}
					className={cn(
						'flex size-6 cursor-pointer items-center justify-center text-dim transition-colors hover:text-foreground',
						theme === value && 'bg-accent text-foreground',
					)}
				>
					<Icon size={12} />
				</button>
			))}
		</div>
	)
}
