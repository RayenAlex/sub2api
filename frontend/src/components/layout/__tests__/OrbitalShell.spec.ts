import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const here = dirname(fileURLToPath(import.meta.url))
const appLayoutSource = readFileSync(resolve(here, '../AppLayout.vue'), 'utf8')
const appHeaderSource = readFileSync(resolve(here, '../AppHeader.vue'), 'utf8')
const appSidebarSource = readFileSync(resolve(here, '../AppSidebar.vue'), 'utf8')
const globalStyleSource = readFileSync(resolve(here, '../../../style.css'), 'utf8')
const tailwindSource = readFileSync(resolve(here, '../../../../tailwind.config.js'), 'utf8')

describe('Orbital telemetry UI framework', () => {
  it('provides a dedicated application shell and bounded telemetry workspace', () => {
    expect(appLayoutSource).toContain('orbital-shell')
    expect(appLayoutSource).toContain('orbital-main')
    expect(appLayoutSource).toContain('orbital-workspace')
    expect(appLayoutSource).toContain('lg:ml-[320px]')
    expect(appLayoutSource).toContain('lg:ml-[88px]')
  })

  it('renders a control-surface header and an aligned sidebar rail', () => {
    expect(appHeaderSource).toContain('orbital-header')
    expect(appHeaderSource).toContain('telemetry-header__title')
    expect(appHeaderSource).toContain('telemetry-header__control')
    expect(appHeaderSource).toContain('telemetry-header__brand-pill')
    expect(appHeaderSource).toContain('telemetry-header__version')
    expect(appHeaderSource).toContain('ORBITAL TELEMETRY CONTROL')
    expect(appHeaderSource).toContain('{{ siteName }}')
    expect(appHeaderSource).toContain(':version="siteVersion"')
    expect(appHeaderSource).toContain('expanded />')
    expect(appHeaderSource).toContain('lg:pl-[31px]')
    expect(appHeaderSource).not.toContain('{{ pageTitle }}')
    expect(globalStyleSource).toContain('max-width: none')
    expect(appSidebarSource).toContain('telemetry-sidebar')
    expect(appSidebarSource).toContain('telemetry-sidebar__surface')
    expect(appSidebarSource).not.toContain('telemetry-sidebar__header')
    expect(appSidebarSource).not.toContain('telemetry-sidebar__mission-meta')
    expect(appSidebarSource).toContain("sidebarCollapsed ? 'w-[88px]' : 'w-[320px]'")
  })

  it('defines neutral surface and telemetry accent tokens', () => {
    expect(globalStyleSource).toContain('--orbital-surface: #faf8fe')
    expect(globalStyleSource).toContain('--orbital-container: #eeedf3')
    expect(globalStyleSource).toContain('.orbital-panel')
    expect(globalStyleSource).toContain('.telemetry-kpi')
    expect(tailwindSource).toContain('telemetry: {')
  })
})


describe('account usage ledger layout', () => {
  it('fills the panel without capping the table at the compact width', () => {
    const rule = globalStyleSource.match(
      /\.telemetry-usage-table--fluid \.data-table--telemetry table\s*\{([^}]+)\}/
    )?.[1]
    expect(rule).toBeDefined()
    expect(rule).toContain('width: 100% !important')
    expect(rule).toContain('min-width: var(--telemetry-layout-width, 1100px) !important')
    expect(rule).toContain('max-width: none')
    expect(rule).toContain('table-layout: auto')
  })

  it('keeps duration and time values unwrapped in the fluid account ledger', () => {
    const rule = globalStyleSource.match(
      /\.telemetry-usage-table--fluid \.data-table--telemetry \[data-column-key='latency'\],[^{]+\{([^}]+)\}/
    )?.[1]
    expect(rule).toContain('white-space: nowrap')
    expect(rule).toContain('text-align: right')
  })
})
