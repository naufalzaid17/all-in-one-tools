import { ref } from 'vue'

type Theme = 'dark' | 'light'

const theme = ref<Theme>(
  document.documentElement.classList.contains('dark') ? 'dark' : 'light',
)

function apply(next: Theme) {
  theme.value = next
  document.documentElement.classList.toggle('dark', next === 'dark')
  try {
    localStorage.setItem('theme', next)
  } catch {
    // Persisting the preference is best-effort.
  }
}

/** Shared dark/light theme state, persisted to localStorage. */
export function useTheme() {
  return {
    theme,
    toggle: () => apply(theme.value === 'dark' ? 'light' : 'dark'),
  }
}
