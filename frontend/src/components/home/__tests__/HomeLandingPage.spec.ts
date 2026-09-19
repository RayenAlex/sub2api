import { describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'

import HomeLandingPage from '../HomeLandingPage.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const baseProps = {
  siteName: 'Test site',
  siteLogo: '',
  siteSubtitle: 'Configured subtitle',
  docUrl: '',
  isAuthenticated: false,
  dashboardPath: '/dashboard',
  showModelPlazaEntry: true,
}

function mountLanding(overrides: Partial<typeof baseProps> = {}) {
  return mount(HomeLandingPage, {
    props: { ...baseProps, ...overrides },
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        LocaleSwitcher: { template: '<div data-testid="locale-switcher" />' },
        Icon: { template: '<span data-testid="icon" />' },
      },
    },
  })
}

describe('HomeLandingPage', () => {
  it('renders the design sections and static marketing content', () => {
    const wrapper = mountLanding()

    expect(wrapper.get('[data-testid="home-landing"]').exists()).toBe(true)
    expect(wrapper.get('#hero').exists()).toBe(true)
    expect(wrapper.get('#heroTerminal').exists()).toBe(true)
    expect(wrapper.get('[data-testid="quick-specs-bar"]').exists()).toBe(true)
    expect(wrapper.get('#models').exists()).toBe(true)
    expect(wrapper.get('#architecture').exists()).toBe(true)
    expect(wrapper.get('#telemetry').exists()).toBe(true)
    expect(wrapper.get('#pricing').exists()).toBe(true)
    expect(wrapper.get('#docs').exists()).toBe(true)
    expect(wrapper.findAll('[data-testid="model-card"]')).toHaveLength(5)
    expect(wrapper.findAll('[data-testid="capability-card"]')).toHaveLength(3)
    expect(wrapper.findAll('[data-testid="telemetry-metric"]')).toHaveLength(4)
    expect(wrapper.get('[data-testid="home-hero-title"]').text()).toContain('home.design.hero.title')
    expect(wrapper.get('[data-testid="home-cta-title"]').text()).toContain('home.design.cta.title')
  })

  it('uses login as the primary destination for anonymous visitors', () => {
    const wrapper = mountLanding()

    const primaryCta = wrapper.findAllComponents(RouterLinkStub).find((link) => link.attributes('data-testid') === 'home-primary-cta')
    const headerAuth = wrapper.findAllComponents(RouterLinkStub).find((link) => link.attributes('data-testid') === 'home-header-auth')

    expect(primaryCta?.props('to')).toBe('/login')
    expect(headerAuth?.props('to')).toBe('/login')
  })

  it('uses the supplied dashboard destination for authenticated visitors', () => {
    const wrapper = mountLanding({
      isAuthenticated: true,
      dashboardPath: '/admin/dashboard',
    })

    const primaryCta = wrapper.findAllComponents(RouterLinkStub).find((link) => link.attributes('data-testid') === 'home-primary-cta')
    const headerAuth = wrapper.findAllComponents(RouterLinkStub).find((link) => link.attributes('data-testid') === 'home-header-auth')

    expect(primaryCta?.props('to')).toBe('/admin/dashboard')
    expect(headerAuth?.props('to')).toBe('/admin/dashboard')
  })

  it('keeps the model plaza entry conditional', () => {
    expect(mountLanding({ showModelPlazaEntry: true }).find('[data-testid="model-plaza-link"]').exists()).toBe(true)
    expect(mountLanding({ showModelPlazaEntry: false }).find('[data-testid="model-plaza-link"]').exists()).toBe(false)
  })

  it('falls back to an in-page documentation anchor when no external URL is configured', () => {
    const wrapper = mountLanding()

    expect(wrapper.get('[data-testid="home-doc-link"]').attributes('href')).toBe('#docs')
  })

  it('uses the configured documentation URL when available', () => {
    const wrapper = mountLanding({ docUrl: 'https://docs.example.com' })

    const link = wrapper.get('[data-testid="home-doc-link"]')
    expect(link.attributes('href')).toBe('https://docs.example.com')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel')).toBe('noopener noreferrer')
  })

  it('initializes coverflow carousel at index 2 and updates on controls/interactions', async () => {
    const wrapper = mountLanding()

    const cards = wrapper.findAll('[data-testid="model-card"]')
    expect(cards).toHaveLength(5)

    // Initial card 2 (GPT) is active
    expect(cards[2].classes()).toContain('bg-white')
    expect(cards[2].classes()).toContain('border-neutral-300')
    expect(cards[2].attributes('style')).toContain('scale(1.05)')
    expect(cards[2].attributes('style')).toContain('z-index: 30')

    // Inactive cards have muted styling
    expect(cards[1].classes()).toContain('bg-neutral-50/95')
    expect(cards[1].attributes('style')).toContain('scale(0.9)')
    expect(cards[1].attributes('style')).toContain('rotateY(24deg)')

    const prevBtn = wrapper.get('#prevModelBtn')
    const nextBtn = wrapper.get('#nextModelBtn')

    // At index 1, neither prev nor next is disabled
    expect(prevBtn.attributes('disabled')).toBeUndefined()
    expect(nextBtn.attributes('disabled')).toBeUndefined()

    // Click next button -> moves to index 3 (Gemini)
    await nextBtn.trigger('click')
    expect(cards[3].classes()).toContain('bg-white')
    expect(cards[3].attributes('style')).toContain('scale(1.05)')
    expect(cards[2].classes()).toContain('bg-neutral-50/95')

    // Click dot 4 -> moves to index 4 (DeepSeek)
    const dots = wrapper.findAll('#coverflowDots button')
    expect(dots).toHaveLength(5)
    await dots[4].trigger('click')
    expect(cards[4].classes()).toContain('bg-white')
    expect(cards[4].attributes('style')).toContain('scale(1.05)')
    expect(nextBtn.attributes('disabled')).toBeDefined()

    // Click side card 0 -> moves to index 0 (Llama)
    await cards[0].trigger('click')
    expect(cards[0].classes()).toContain('bg-white')
    expect(cards[0].attributes('style')).toContain('scale(1.05)')
    expect(prevBtn.attributes('disabled')).toBeDefined()

    // Keyboard ArrowRight on track -> moves to index 1
    const track = wrapper.get('#coverflowTrack')
    await track.trigger('keydown', { key: 'ArrowRight' })
    expect(cards[1].classes()).toContain('bg-white')
  })

  it('reroutes dead-zone clicks on the center card to the owning side card', async () => {
    const wrapper = mountLanding()
    const cards = wrapper.findAll('[data-testid="model-card"]')
    const track = wrapper.get('#coverflowTrack')

    // jsdom has no layout: give the track a fake 1280px rect so zone math runs.
    vi.spyOn(track.element, 'getBoundingClientRect').mockReturnValue({
      left: 0,
      right: 1280,
      width: 1280,
      top: 0,
      bottom: 440,
      height: 440,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    } as DOMRect)

    // Click lands on the center card's surface (dead zone) but inside card 3's zone.
    await cards[2].trigger('click', { clientX: 640 + 300 })
    expect(cards[3].classes()).toContain('bg-white')
    expect(cards[2].classes()).toContain('bg-neutral-50/95')

    // Click inside the center zone on the center card -> no move.
    await cards[3].trigger('click', { clientX: 640 + 10 })
    expect(cards[3].classes()).toContain('bg-white')
    expect(cards[1].classes()).toContain('bg-neutral-50/95')
  })

  it('renders Apple-style scroll reveal structures and stagger attributes across sections', () => {
    const wrapper = mountLanding()

    // Hero title and CTA have reveal classes
    const heroTitle = wrapper.get('[data-testid="home-hero-title"]')
    expect(heroTitle.classes()).toContain('reveal-fade-up')
    expect(heroTitle.attributes('style')).toContain('--i: 1')

    const heroCta = wrapper.get('[data-testid="home-primary-cta"]').element.parentElement
    expect(heroCta?.classList.contains('reveal-pop')).toBe(true)

    // QuickSpecsBar has 4 staggered items
    const quickSpecsBar = wrapper.get('[data-testid="quick-specs-bar"]')
    expect(quickSpecsBar.classes()).toContain('reveal-group')
    const quickSpecItems = quickSpecsBar.findAll('.reveal-fade-up')
    expect(quickSpecItems).toHaveLength(4)
    expect(quickSpecItems[0].attributes('style')).toContain('--i: 0')
    expect(quickSpecItems[3].attributes('style')).toContain('--i: 3')

    // Model matrix section has reveal-group and reveals
    const modelMatrix = wrapper.get('[data-testid="home-model-matrix"]')
    expect(modelMatrix.classes()).toContain('reveal-group')
    expect(modelMatrix.findAll('.reveal-fade-up').length).toBeGreaterThan(0)

    // Capabilities and telemetry sections have reveal-group
    expect(wrapper.get('[data-testid="home-capabilities"]').classes()).toContain('reveal-group')
    expect(wrapper.get('[data-testid="home-telemetry"]').classes()).toContain('reveal-group')
  })
})
