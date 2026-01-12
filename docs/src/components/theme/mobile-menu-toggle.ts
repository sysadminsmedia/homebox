import type { BaseElement } from '@aria-ui/core'
import {
  createSignal,
  defineCustomElement,
  useAriaAttribute,
  useEffect,
  useEventListener,
} from '@aria-ui/core'

function setupMobileMenuToggle(host: BaseElement) {
  const expanded = createSignal(false)

  function toggleExpanded() {
    expanded.set(!expanded.get())
  }

  useAriaAttribute(host, 'aria-expanded', () => {
    return expanded.get() ? 'true' : 'false'
  })

  useEffect(host, () => {
    document.body.toggleAttribute(
      'data-mobile-menu-expanded',
      expanded.get() || false,
    )
  })

  // Keep icon visibility in sync with `expanded` using Tailwind `hidden` class.
  useEffect(host, () => {
    const iconOpen = host.querySelector('.icon-open') as HTMLElement | null
    const iconClose = host.querySelector('.icon-close') as HTMLElement | null
    if (!iconOpen || !iconClose) return

    // set visibility based on current state
    iconOpen.classList.toggle('hidden', expanded.get())
    iconClose.classList.toggle('hidden', !expanded.get())
  })

  useEventListener(host, 'click', toggleExpanded)
  useEventListener(host, 'keydown', (event) => {
    if (!event.isComposing && (event.key === 'Enter' || event.key === ' ')) {
      event.preventDefault()
      toggleExpanded()
    }
  })

  useEffect(host, () => {
    const sidebarNav = document.getElementById('nova-sidebar-nav')
    if (!sidebarNav) return
    const handleKeyup = (event: KeyboardEvent) => {
      if (event.code === 'Escape') {
        expanded.set(false)
        host.focus()
      }
    }
    sidebarNav.addEventListener('keyup', handleKeyup)
    return () => sidebarNav.removeEventListener('keyup', handleKeyup)
  })

  useEffect(host, () => {
    host.setAttribute('role', 'button')
    host.setAttribute('tabindex', '0')
  })
}

const MobileMenuToggle = defineCustomElement({
  props: {},
  events: {},
  setup: setupMobileMenuToggle,
})

function register() {
  customElements.define('nova-mobile-menu-toggle', MobileMenuToggle)
}

export { register }
