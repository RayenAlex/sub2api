import { nextTick, onBeforeUnmount, onMounted, type Directive, type Ref } from 'vue'

export interface ScrollRevealOptions {
  /**
   * Root margin string for IntersectionObserver.
   * Default '0px 0px -25% 0px' triggers when element top reaches ~75% of viewport height.
   */
  rootMargin?: string
  /**
   * Threshold for IntersectionObserver.
   */
  threshold?: number | number[]
  /**
   * CSS selector matching elements to reveal.
   */
  revealSelector?: string
  /**
   * CSS class added when element enters viewport.
   */
  activeClass?: string
  /**
   * Optional scoped container ref.
   */
  scopeRef?: Ref<HTMLElement | null>
}

/**
 * Apple-style scroll reveal composable.
 * Uses IntersectionObserver to trigger entry animations once when scrolled into view (~70% of viewport).
 * Supports --i stagger variables for staggered lists, and respects prefers-reduced-motion.
 */
export function useScrollReveal(
  scopeOrOptions?: Ref<HTMLElement | null> | ScrollRevealOptions,
  maybeOptions: ScrollRevealOptions = {}
) {
  const options: ScrollRevealOptions =
    scopeOrOptions && typeof scopeOrOptions === 'object' && 'value' in scopeOrOptions
      ? { ...maybeOptions, scopeRef: scopeOrOptions as Ref<HTMLElement | null> }
      : (scopeOrOptions as ScrollRevealOptions) || {}

  const {
    rootMargin = '0px 0px -25% 0px',
    threshold = 0.05,
    revealSelector = '.reveal-fade-up, .reveal-pop, .reveal-group',
    activeClass = 'is-revealed',
    scopeRef,
  } = options

  let observer: IntersectionObserver | null = null

  const isReducedMotion = (): boolean => {
    if (typeof window === 'undefined' || !window.matchMedia) return false
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches
  }

  const revealElement = (el: Element) => {
    el.classList.add(activeClass)
    if (el.classList.contains('reveal-group')) {
      const children = el.querySelectorAll('.reveal-fade-up, .reveal-pop')
      children.forEach((child) => child.classList.add(activeClass))
    }
  }

  const init = () => {
    if (typeof window === 'undefined') return

    const container = scopeRef?.value || document.querySelector('.home-landing') || document.body
    if (!container) return

    // Add .reveal-ready to root scope to activate initial hidden states in CSS without SSR flash
    container.classList.add('reveal-ready')

    const allElements = Array.from(container.querySelectorAll(revealSelector))
    if (allElements.length === 0) return

    // If reduced motion is requested or IntersectionObserver is not supported, reveal all immediately
    if (isReducedMotion() || typeof IntersectionObserver === 'undefined') {
      allElements.forEach(revealElement)
      return
    }

    // Direct targets to observe: if an element is inside a .reveal-group and not the group itself,
    // the parent .reveal-group handles triggering its children to guarantee coordinated stagger timing.
    const targetsToObserve = allElements.filter((el) => {
      if (el.classList.contains('reveal-group')) return true
      const parentGroup = el.parentElement?.closest('.reveal-group')
      return !parentGroup
    })

    try {
      observer = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (entry.isIntersecting) {
              revealElement(entry.target)
              observer?.unobserve(entry.target)
            }
          })
        },
        {
          root: null,
          rootMargin,
          threshold,
        }
      )

      targetsToObserve.forEach((el) => {
        observer?.observe(el)
      })
    } catch {
      // Fallback
      allElements.forEach(revealElement)
    }
  }

  const cleanup = () => {
    if (observer) {
      observer.disconnect()
      observer = null
    }
  }

  onMounted(async () => {
    await nextTick()
    init()
  })

  onBeforeUnmount(() => {
    cleanup()
  })

  return {
    init,
    cleanup,
    revealElement,
  }
}

/**
 * Optional Vue directive `v-reveal` or `v-reveal:pop="index"`.
 */
export const vReveal: Directive = {
  mounted(el: HTMLElement, binding) {
    if (typeof window === 'undefined') return
    if (binding.arg === 'pop') {
      el.classList.add('reveal-pop')
    } else {
      el.classList.add('reveal-fade-up')
    }
    if (binding.value !== undefined) {
      el.style.setProperty('--i', String(binding.value))
    }
  },
}
