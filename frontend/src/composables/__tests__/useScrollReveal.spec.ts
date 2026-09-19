import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { useScrollReveal } from '../useScrollReveal'

describe('useScrollReveal', () => {
  let originalMatchMedia: typeof window.matchMedia
  let originalIntersectionObserver: typeof window.IntersectionObserver

  beforeEach(() => {
    originalMatchMedia = window.matchMedia
    originalIntersectionObserver = window.IntersectionObserver

    window.matchMedia = vi.fn().mockImplementation((query) => ({
      matches: query !== '(prefers-reduced-motion: reduce)',
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }))
  })

  afterEach(() => {
    window.matchMedia = originalMatchMedia
    window.IntersectionObserver = originalIntersectionObserver
    globalThis.IntersectionObserver = originalIntersectionObserver
    vi.restoreAllMocks()
  })

  it('initializes and observes elements in standard mode', async () => {
    let observedElements: Element[] = []
    const observeMock = vi.fn((el: Element) => observedElements.push(el))
    const disconnectMock = vi.fn()
    const unobserveMock = vi.fn()

    class TestIntersectionObserver {
      constructor(public callback: IntersectionObserverCallback) {}
      observe = observeMock
      disconnect = disconnectMock
      unobserve = unobserveMock
    }

    window.IntersectionObserver = TestIntersectionObserver as unknown as typeof IntersectionObserver
    globalThis.IntersectionObserver = window.IntersectionObserver

    const TestComp = defineComponent({
      setup() {
        const rootRef = ref<HTMLElement | null>(null)
        useScrollReveal({ scopeRef: rootRef })
        return () =>
          h('div', { ref: rootRef, class: 'test-root' }, [
            h('div', { class: 'reveal-fade-up' }, 'Item 1'),
            h('div', { class: 'reveal-pop' }, 'Item 2'),
          ])
      },
    })

    const wrapper = mount(TestComp)
    await nextTick()

    expect(wrapper.classes()).toContain('reveal-ready')
    expect(observeMock).toHaveBeenCalledTimes(2)

    wrapper.unmount()
    expect(disconnectMock).toHaveBeenCalled()
  })

  it('reveals element and group children when intersecting', async () => {
    let capturedCallback: IntersectionObserverCallback | null = null
    const unobserveMock = vi.fn()

    class TestIntersectionObserver {
      constructor(public callback: IntersectionObserverCallback) {
        capturedCallback = callback
      }
      observe = vi.fn()
      disconnect = vi.fn()
      unobserve = unobserveMock
    }

    window.IntersectionObserver = TestIntersectionObserver as unknown as typeof IntersectionObserver
    globalThis.IntersectionObserver = window.IntersectionObserver

    const TestComp = defineComponent({
      setup() {
        const rootRef = ref<HTMLElement | null>(null)
        useScrollReveal({ scopeRef: rootRef })
        return () =>
          h('div', { ref: rootRef, class: 'test-root' }, [
            h('div', { class: 'reveal-group' }, [
              h('div', { class: 'reveal-fade-up' }, 'Child 1'),
              h('div', { class: 'reveal-pop' }, 'Child 2'),
            ]),
          ])
      },
    })

    const wrapper = mount(TestComp)
    await nextTick()
    const groupEl = wrapper.find('.reveal-group').element

    expect(capturedCallback).not.toBeNull()
    capturedCallback!(
      [
        {
          isIntersecting: true,
          target: groupEl,
        } as unknown as IntersectionObserverEntry,
      ],
      {} as IntersectionObserver
    )

    expect(groupEl.classList.contains('is-revealed')).toBe(true)
    const child1 = wrapper.find('.reveal-fade-up').element
    const child2 = wrapper.find('.reveal-pop').element
    expect(child1.classList.contains('is-revealed')).toBe(true)
    expect(child2.classList.contains('is-revealed')).toBe(true)
    expect(unobserveMock).toHaveBeenCalledWith(groupEl)
  })

  it('immediately reveals elements and skips observer if prefers-reduced-motion matches', async () => {
    const observeMock = vi.fn()
    class TestIntersectionObserver {
      observe = observeMock
      disconnect = vi.fn()
      unobserve = vi.fn()
    }
    window.IntersectionObserver = TestIntersectionObserver as unknown as typeof IntersectionObserver
    globalThis.IntersectionObserver = window.IntersectionObserver

    window.matchMedia = vi.fn().mockImplementation((query) => ({
      matches: query === '(prefers-reduced-motion: reduce)',
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }))

    const TestComp = defineComponent({
      setup() {
        const rootRef = ref<HTMLElement | null>(null)
        useScrollReveal({ scopeRef: rootRef })
        return () =>
          h('div', { ref: rootRef, class: 'test-root' }, [
            h('div', { class: 'reveal-fade-up' }, 'Item 1'),
            h('div', { class: 'reveal-pop' }, 'Item 2'),
          ])
      },
    })

    const wrapper = mount(TestComp)
    await nextTick()

    expect(wrapper.find('.reveal-fade-up').classes()).toContain('is-revealed')
    expect(wrapper.find('.reveal-pop').classes()).toContain('is-revealed')
    expect(observeMock).not.toHaveBeenCalled()
  })
})
