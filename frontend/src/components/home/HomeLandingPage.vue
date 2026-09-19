<template>
  <div ref="homeLandingRef" data-testid="home-landing" class="home-landing min-h-screen overflow-x-hidden bg-white text-[#1d1d1f]">
    <header class="sticky top-0 z-40 px-4 pt-4 sm:px-6 lg:px-8">
      <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 rounded-full border border-[#eeedf3] bg-white/90 px-3 py-2 shadow-[0_2px_12px_rgba(0,0,0,0.04)] backdrop-blur-xl">
        <router-link to="/home" class="flex min-w-0 shrink-0 items-center gap-2 rounded-full px-2 py-1.5" aria-label="Home">
          <img :src="siteLogo || '/logo.svg'" :alt="siteName" class="h-8 w-8 rounded-lg object-contain" />
          <span class="hidden max-w-32 truncate text-sm font-semibold sm:inline">{{ siteName }}</span>
        </router-link>

        <nav aria-label="Homepage sections" class="hidden items-center gap-1 rounded-full bg-[#f4f3f8] p-1 md:flex">
          <a
            v-for="link in navigationLinks"
            :key="link.href"
            :href="link.href"
            class="rounded-full px-3 py-2 text-xs font-medium text-[#77767b] transition-colors hover:bg-white hover:text-[#1b1b1f] lg:px-4"
          >
            {{ link.label }}
          </a>
        </nav>

        <div class="flex shrink-0 items-center gap-1.5 sm:gap-2">
          <LocaleSwitcher />
          <router-link
            v-if="showModelPlazaEntry"
            data-testid="model-plaza-link"
            to="/model-plaza"
            class="hidden rounded-full px-3 py-2 text-xs font-medium text-[#77767b] transition-colors hover:bg-[#f4f3f8] hover:text-[#1b1b1f] lg:inline-flex"
          >
            {{ t('nav.modelPlaza') }}
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="hidden h-9 w-9 items-center justify-center rounded-full text-[#77767b] transition-colors hover:bg-[#f4f3f8] hover:text-[#1b1b1f] sm:inline-flex"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="sm" />
          </a>
          <router-link
            data-testid="home-header-auth"
            :to="primaryDestination"
            class="inline-flex items-center rounded-full bg-[#030304] px-3.5 py-2 text-xs font-semibold text-white transition-transform hover:-translate-y-0.5 hover:bg-[#303034] active:scale-[0.98] sm:px-4"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </div>
    </header>

    <main>
      <section id="hero" class="relative overflow-hidden pb-20 pt-32 md:pb-28 md:pt-44">
        <!-- Background Tesla-style soft gradient spotlight -->
        <div
          ref="heroHaloRef"
          class="pointer-events-none absolute left-1/2 top-0 -z-10 h-[550px] w-[1000px] -translate-x-1/2 bg-gradient-to-b from-neutral-100 via-neutral-50/50 to-transparent blur-3xl will-change-transform"
        ></div>
        <div class="mx-auto max-w-6xl px-6 text-center">
          <!-- Subtitle badge pill -->
          <div
            class="reveal-pop mb-8 inline-flex items-center gap-2 rounded-full border border-neutral-200/90 bg-neutral-50/80 px-3.5 py-1 text-[13px] font-medium text-neutral-600"
            style="--i: 0"
          >
            <span class="h-2 w-2 animate-pulse rounded-full bg-emerald-500"></span>
            <span>{{ t('home.design.hero.eyebrow') }}</span>
            <span class="text-neutral-300">|</span>
            <span class="text-neutral-800">{{ t('home.design.hero.status') }}</span>
          </div>

          <!-- Grand Statement Title -->
          <h1
            data-testid="home-hero-title"
            class="reveal-fade-up mx-auto mb-7 max-w-4xl text-balance text-center text-5xl font-semibold leading-[1.06] tracking-[-0.035em] text-neutral-950 sm:text-6xl md:text-7xl lg:text-[82px]"
            style="--i: 1"
            v-html="t('home.design.hero.title')"
          ></h1>

          <!-- Optional siteSubtitle prop -->
          <p
            v-if="siteSubtitle"
            class="reveal-fade-up mx-auto -mt-3 mb-6 max-w-2xl text-balance text-center text-base font-medium leading-relaxed text-neutral-600 sm:text-lg"
            style="--i: 2"
          >
            {{ siteSubtitle }}
          </p>

          <!-- Subheading -->
          <p
            class="reveal-fade-up mx-auto mb-10 max-w-2xl text-balance text-center text-lg font-normal leading-relaxed text-neutral-500 md:text-xl"
            :style="{ '--i': siteSubtitle ? 3 : 2 }"
          >
            {{ t('home.design.hero.description') }}
          </p>

          <!-- CTA Buttons -->
          <div
            class="reveal-pop mb-16 flex flex-col items-center justify-center gap-4 sm:flex-row"
            :style="{ '--i': siteSubtitle ? 4 : 3 }"
          >
            <router-link
              data-testid="home-primary-cta"
              :to="primaryDestination"
              class="flex w-full items-center justify-center gap-2 rounded-full bg-black px-8 py-3.5 text-sm font-medium text-white shadow-lg shadow-neutral-900/10 transition-all duration-200 hover:bg-neutral-800 hover:shadow-neutral-900/20 sm:w-auto"
            >
              <span>{{ t('home.design.hero.primary') }}</span>
              <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path d="M14 5l7 7m0 0l-7 7m7-7H3" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </router-link>
            <a
              href="#models"
              class="flex w-full items-center justify-center rounded-full bg-neutral-100 px-8 py-3.5 text-sm font-medium text-neutral-800 transition-all duration-200 hover:bg-neutral-200/80 sm:w-auto"
            >
              <span>{{ t('home.design.hero.secondary') }}</span>
            </a>
          </div>

          <!-- Hero Showcase Console & Live Spec -->
          <div
            class="hero-terminal-wrap reveal-fade-up relative mx-auto mt-6 max-w-4xl"
            :style="{ '--i': siteSubtitle ? 5 : 4 }"
          >
            <div
              id="heroTerminal"
              ref="heroTerminalRef"
              class="dark-console-bg rounded-2xl border border-neutral-800/80 p-5 text-left font-mono shadow-2xl md:p-7 will-change-transform"
              style="transform: scale(0.98) rotateX(2.2deg);"
            >
              <!-- Window Bar -->
              <div class="mb-5 flex items-center justify-between border-b border-neutral-800/90 pb-4 text-xs text-neutral-500">
                <div class="flex items-center gap-2">
                  <span class="inline-block h-3 w-3 rounded-full bg-[#ff5f56]"></span>
                  <span class="inline-block h-3 w-3 rounded-full bg-[#ffbd2e]"></span>
                  <span class="inline-block h-3 w-3 rounded-full bg-[#27c93f]"></span>
                  <span class="ml-3 font-sans font-medium text-neutral-400">{{ t('home.design.hero.terminalTitle') }}</span>
                </div>
                <div class="hidden items-center gap-4 font-mono text-[11px] text-neutral-400 sm:flex">
                  <span>{{ t('home.design.hero.terminalProtocol') }}</span>
                  <span class="flex items-center gap-1.5 text-emerald-400">
                    <span class="inline-block h-1.5 w-1.5 animate-ping rounded-full bg-emerald-400"></span>
                    {{ t('home.design.hero.terminalActive') }}
                  </span>
                </div>
              </div>

              <!-- Code Execution Stream -->
              <div class="space-y-3 text-xs leading-relaxed md:text-[13px]">
                <div class="text-neutral-400">
                  <span class="font-bold text-emerald-400">$</span> curl -X POST https://api.sub2.org/v1/chat/completions \
                </div>
                <div class="pl-4 text-neutral-500">
                  -H <span class="text-amber-200">"Authorization: Bearer sk-sub2-live-production"</span> \<br />
                  -H <span class="text-amber-200">"Content-Type: application/json"</span> \<br />
                  -d <span class="text-cyan-300">'{ "model": "claude-3-7-sonnet", "stream": true, "messages": [...] }'</span>
                </div>
                <div class="flex items-center gap-2 pt-2 text-xs text-neutral-400">
                  <span class="inline-block rounded border border-emerald-800 bg-emerald-950 px-1.5 py-0.5 text-[11px] font-semibold text-emerald-300">{{ t('home.design.hero.terminalSuccess') }}</span>
                  <span class="text-neutral-400">{{ t('home.design.hero.terminalRouting') }}</span>
                </div>
                <div class="rounded-lg border border-neutral-800 bg-neutral-900/90 p-3.5 font-mono text-[12px] leading-normal text-neutral-200">
                  <span class="text-neutral-500">&gt; data:</span> {"id":"chatcmpl-9x","choices":[{"delta":{"content":"<span class="text-emerald-300" id="streamingText">{{ displayedStreamingText }}</span><span class="stream-cursor"></span>"}}]}
                </div>
              </div>

              <!-- Realtime Telemetry Grid Inside Card -->
              <div class="mt-6 grid grid-cols-2 gap-4 border-t border-neutral-800/80 pt-5 text-left font-sans sm:grid-cols-4">
                <div>
                  <div class="text-[11px] font-medium uppercase tracking-wider text-neutral-500">{{ t('home.design.hero.terminalTelemetry.latency.label') }}</div>
                  <div class="mt-0.5 text-lg font-bold tracking-tight text-white md:text-xl">{{ t('home.design.hero.terminalTelemetry.latency.value') }}</div>
                  <div class="mt-0.5 text-[11px] text-emerald-400">{{ t('home.design.hero.terminalTelemetry.latency.note') }}</div>
                </div>
                <div>
                  <div class="text-[11px] font-medium uppercase tracking-wider text-neutral-500">{{ t('home.design.hero.terminalTelemetry.healing.label') }}</div>
                  <div class="mt-0.5 text-lg font-bold tracking-tight text-white md:text-xl">{{ t('home.design.hero.terminalTelemetry.healing.value') }}</div>
                  <div class="mt-0.5 text-[11px] text-neutral-400">{{ t('home.design.hero.terminalTelemetry.healing.note') }}</div>
                </div>
                <div>
                  <div class="text-[11px] font-medium uppercase tracking-wider text-neutral-500">{{ t('home.design.hero.terminalTelemetry.overhead.label') }}</div>
                  <div class="mt-0.5 text-lg font-bold tracking-tight text-white md:text-xl">{{ t('home.design.hero.terminalTelemetry.overhead.value') }}</div>
                  <div class="mt-0.5 text-[11px] text-neutral-400">{{ t('home.design.hero.terminalTelemetry.overhead.note') }}</div>
                </div>
                <div>
                  <div class="text-[11px] font-medium uppercase tracking-wider text-neutral-500">{{ t('home.design.hero.terminalTelemetry.affinity.label') }}</div>
                  <div class="mt-0.5 text-lg font-bold tracking-tight text-white md:text-xl">{{ t('home.design.hero.terminalTelemetry.affinity.value') }}</div>
                  <div class="mt-0.5 text-[11px] text-emerald-400">{{ t('home.design.hero.terminalTelemetry.affinity.note') }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- BEGIN: QuickSpecsBar -->
      <section data-testid="quick-specs-bar" class="reveal-group border-y border-neutral-200/70 bg-[#f5f5f7] py-5">
        <div class="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-x-8 gap-y-4 px-6 text-xs font-medium text-neutral-600 md:text-sm">
          <div class="reveal-fade-up flex items-center gap-2.5" style="--i: 0">
            <span class="h-2 w-2 rounded-full bg-black"></span>
            <span>{{ t('home.design.quickSpecs.automated') }}</span>
          </div>
          <div class="reveal-fade-up flex items-center gap-2.5" style="--i: 1">
            <span class="h-2 w-2 rounded-full bg-black"></span>
            <span>{{ t('home.design.quickSpecs.context') }}</span>
          </div>
          <div class="reveal-fade-up flex items-center gap-2.5" style="--i: 2">
            <span class="h-2 w-2 rounded-full bg-black"></span>
            <span>{{ t('home.design.quickSpecs.billing') }}</span>
          </div>
          <div class="reveal-fade-up flex items-center gap-2.5" style="--i: 3">
            <span class="h-2 w-2 rounded-full bg-black"></span>
            <span>{{ t('home.design.quickSpecs.sdk') }}</span>
          </div>
        </div>
      </section>
      <!-- END: QuickSpecsBar -->

      <section id="models" data-testid="home-model-matrix" class="reveal-group relative scroll-mt-28 overflow-hidden border-y border-[#eeedf3] bg-white px-4 py-24 sm:px-6 lg:px-8 lg:py-32">
        <div class="mx-auto max-w-6xl">
          <div class="mb-12 flex flex-col justify-between gap-4 md:flex-row md:items-end">
            <div class="reveal-fade-up" style="--i: 0">
              <p class="mb-2 text-xs font-semibold uppercase tracking-widest text-neutral-400">{{ t('home.design.models.eyebrow') }}</p>
              <h2 class="text-3xl font-semibold tracking-tight text-neutral-900 sm:text-4xl">{{ t('home.design.models.title') }}</h2>
            </div>
            <div class="flex items-center gap-4">
              <p class="reveal-fade-up hidden max-w-md text-sm font-normal text-neutral-500 md:block" style="--i: 1">{{ t('home.design.models.description') }}</p>
              <div class="reveal-pop flex items-center gap-2" style="--i: 2">
                <button
                  id="prevModelBtn"
                  type="button"
                  :disabled="activeIndex === 0"
                  aria-label="Previous model"
                  class="flex h-9 w-9 items-center justify-center rounded-full border border-neutral-200 bg-neutral-50/80 text-neutral-700 shadow-sm transition-all hover:bg-neutral-100 active:scale-95 disabled:pointer-events-none disabled:opacity-30"
                  @click="goTo(activeIndex - 1)"
                >
                  <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path d="M15 19l-7-7 7-7" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </button>
                <button
                  id="nextModelBtn"
                  type="button"
                  :disabled="activeIndex === modelCards.length - 1"
                  aria-label="Next model"
                  class="flex h-9 w-9 items-center justify-center rounded-full border border-neutral-200 bg-neutral-50/80 text-neutral-700 shadow-sm transition-all hover:bg-neutral-100 active:scale-95 disabled:pointer-events-none disabled:opacity-30"
                  @click="goTo(activeIndex + 1)"
                >
                  <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path d="M9 5l7 7-7 7" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
          <p class="reveal-fade-up mb-6 text-sm font-normal text-neutral-500 md:hidden" style="--i: 1">{{ t('home.design.models.description') }}</p>

          <div class="reveal-fade-up relative w-full" style="--i: 2">
            <div class="carousel-perspective-viewport relative w-full py-8">
              <div class="pointer-events-none absolute left-1/2 top-1/2 -z-10 h-72 w-80 -translate-x-1/2 -translate-y-1/2 rounded-full bg-gradient-to-r from-blue-100/40 via-purple-100/30 to-amber-100/40 blur-3xl sm:w-[480px]"></div>
              <div
                id="coverflowTrack"
                ref="trackEl"
                tabindex="0"
                class="relative mx-auto flex h-[440px] w-full items-center justify-center focus:outline-none sm:h-[410px]"
                @keydown.left.prevent="goTo(activeIndex - 1)"
                @keydown.right.prevent="goTo(activeIndex + 1)"
                @touchstart.passive="handleTouchStart"
                @touchend.passive="handleTouchEnd"
                @click="handleTrackClick"
              >
                <div
                  v-for="(model, index) in modelCards"
                  :key="model.key"
                  data-testid="model-card"
                  :data-index="index"
                  class="coverflow-card group flex h-[380px] cursor-pointer flex-col justify-between rounded-3xl border p-7 backdrop-blur-sm sm:h-[350px]"
                  :class="index === activeIndex ? 'border-neutral-300 bg-white' : 'border-neutral-200/90 bg-neutral-50/95'"
                  :style="getCardStyle(index)"
                  @click="handleCardClick(index, $event)"
                >
                  <div>
                    <div class="mb-4 flex items-center justify-between">
                      <div
                        class="flex h-11 w-11 items-center justify-center rounded-2xl text-lg font-bold text-white shadow-sm transition-transform duration-300 group-hover:scale-105"
                        :class="model.markClass"
                      >
                        {{ model.mark }}
                      </div>
                      <span
                        class="rounded-full px-2.5 py-0.5 text-[11px] font-semibold"
                        :class="model.badgeClass"
                      >
                        {{ model.badge }}
                      </span>
                    </div>
                    <h3 class="text-xl font-semibold tracking-tight text-neutral-900">{{ model.name }}</h3>
                    <p class="mb-4 font-mono text-xs text-neutral-500">{{ model.vendor }}</p>
                    <p class="mb-6 text-xs leading-relaxed text-neutral-600 md:text-[13px]">
                      {{ model.description }}
                    </p>
                  </div>
                  <div class="flex items-center justify-between border-t border-neutral-200/70 pt-4">
                    <span class="font-mono text-[11px] font-medium text-neutral-500">{{ model.note }}</span>
                    <a
                      :href="docHref"
                      :target="docTarget"
                      :rel="docRel"
                      class="inline-flex items-center gap-1 text-xs font-semibold text-neutral-900 transition-colors group-hover:text-[#0071e3]"
                    >
                      {{ t('home.design.models.guide') }}
                    </a>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="reveal-pop mt-8 flex items-center justify-center gap-4" style="--i: 3">
            <div id="coverflowDots" class="flex items-center gap-2.5 rounded-full bg-[#e8e8ed] px-6 py-2.5 shadow-inner">
              <button
                v-for="(_, index) in modelCards"
                :key="index"
                type="button"
                :aria-label="`Slide ${index + 1}`"
                :class="[
                  'transition-all duration-300',
                  index === activeIndex ? 'h-2 w-8 rounded-full bg-[#1d1d1f]' : 'h-2 w-2 rounded-full bg-neutral-400/50 hover:bg-neutral-700',
                ]"
                @click="goTo(index)"
              />
            </div>
          </div>
        </div>
      </section>

      <section id="architecture" data-testid="home-capabilities" class="reveal-group scroll-mt-28 bg-[#f4f3f8] px-4 py-24 sm:px-6 lg:px-8 lg:py-32">
        <div class="mx-auto max-w-7xl">
          <div class="max-w-3xl">
            <p class="reveal-fade-up text-xs font-semibold uppercase tracking-[0.18em] text-[#77767b]" style="--i: 0">{{ t('home.design.capabilities.eyebrow') }}</p>
            <h2 class="reveal-fade-up mt-3 text-4xl font-semibold tracking-[-0.05em] text-[#030304] sm:text-5xl" style="--i: 1">{{ t('home.design.capabilities.title') }}</h2>
            <p class="reveal-fade-up mt-5 max-w-2xl text-base leading-7 text-[#77767b]" style="--i: 2">{{ t('home.design.capabilities.description') }}</p>
          </div>

          <div class="mt-12 grid gap-5 lg:grid-cols-3">
            <article
              v-for="(capability, index) in capabilityCards"
              :key="capability.key"
              data-testid="capability-card"
              class="reveal-fade-up flex min-h-[22rem] flex-col rounded-3xl border border-[#eeedf3] bg-white p-7 shadow-[0_2px_12px_rgba(0,0,0,0.03)]"
              :style="{ '--i': index + 2 }"
            >
              <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-[#f4f3f8] text-[#030304]"><Icon :name="capability.icon" size="md" :stroke-width="1.8" /></div>
              <h3 class="mt-8 text-xl font-semibold tracking-[-0.03em] text-[#030304]">{{ capability.title }}</h3>
              <p class="mt-4 text-sm leading-6 text-[#77767b]">{{ capability.description }}</p>
              <p class="mt-auto border-t border-[#eeedf3] pt-6 text-xs font-medium text-[#99989e]">{{ capability.note }}</p>
            </article>
          </div>
        </div>
      </section>

      <section id="telemetry" data-testid="home-telemetry" class="reveal-group scroll-mt-28 bg-white px-4 py-24 sm:px-6 lg:px-8 lg:py-32">
        <div class="mx-auto max-w-7xl">
          <div class="text-center">
            <div class="reveal-pop inline-flex items-center gap-2 rounded-full border border-[#d8f3e5] bg-[#f3fff8] px-3.5 py-2 text-xs font-medium text-[#009a3b]" style="--i: 0">
              <span class="h-2 w-2 animate-pulse rounded-full bg-[#10b981]" aria-hidden="true"></span>
              {{ t('home.design.telemetry.eyebrow') }}
            </div>
            <h2 class="reveal-fade-up mt-5 text-4xl font-semibold tracking-[-0.05em] text-[#030304] sm:text-5xl" style="--i: 1">{{ t('home.design.telemetry.title') }}</h2>
            <p class="reveal-fade-up mx-auto mt-4 max-w-2xl text-sm leading-6 text-[#77767b]" style="--i: 2">{{ t('home.design.telemetry.description') }}</p>
          </div>

          <div class="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <article
              v-for="(metric, index) in telemetryMetrics"
              :key="metric.key"
              data-testid="telemetry-metric"
              class="reveal-fade-up rounded-3xl border border-[#eeedf3] bg-[#faf8fe] px-6 py-8 text-center shadow-[0_2px_12px_rgba(0,0,0,0.03)]"
              :style="{ '--i': index + 2 }"
            >
              <div class="font-mono text-4xl font-semibold tracking-[-0.08em] text-[#030304] sm:text-5xl">{{ metric.value }}</div>
              <p class="mt-3 text-xs font-medium leading-5 text-[#77767b]">{{ metric.label }}</p>
            </article>
          </div>
        </div>
      </section>

      <section id="pricing" data-testid="home-pricing" class="reveal-group scroll-mt-28 border-y border-[#eeedf3] bg-[#faf8fe] px-4 py-24 sm:px-6 lg:px-8 lg:py-32">
        <div class="mx-auto grid max-w-7xl items-center gap-12 lg:grid-cols-[0.9fr_1.1fr] lg:gap-20">
          <div>
            <p class="reveal-fade-up text-xs font-semibold uppercase tracking-[0.18em] text-[#77767b]" style="--i: 0">{{ t('home.design.pricing.eyebrow') }}</p>
            <h2 class="reveal-fade-up mt-3 text-4xl font-semibold tracking-[-0.05em] text-[#030304] sm:text-5xl" style="--i: 1">{{ t('home.design.pricing.title') }}</h2>
            <p class="reveal-fade-up mt-5 max-w-xl text-base leading-7 text-[#77767b]" style="--i: 2">{{ t('home.design.pricing.description') }}</p>
          </div>
          <div class="reveal-fade-up rounded-3xl border border-[#eeedf3] bg-white p-6 shadow-[0_2px_12px_rgba(0,0,0,0.03)] sm:p-8" style="--i: 2">
            <div class="flex items-center justify-between border-b border-[#eeedf3] pb-5">
              <span class="text-sm font-semibold text-[#030304]">{{ t('home.design.pricing.value') }}</span>
              <span class="rounded-full bg-[#e6f8ee] px-3 py-1 text-xs font-semibold text-[#009a3b]">0.00 ¥</span>
            </div>
            <div class="mt-6 grid gap-4 sm:grid-cols-3">
              <div v-for="item in pricingItems" :key="item.key" class="rounded-2xl bg-[#f4f3f8] p-4">
                <div class="font-mono text-lg font-semibold text-[#030304]">{{ item.value }}</div>
                <div class="mt-2 text-xs text-[#77767b]">{{ item.label }}</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="docs" data-testid="home-docs" class="reveal-group scroll-mt-28 bg-white px-4 py-24 sm:px-6 lg:px-8 lg:py-32">
        <div class="mx-auto flex max-w-7xl flex-col gap-8 rounded-3xl border border-[#eeedf3] bg-[#f4f3f8] p-8 sm:p-12 lg:flex-row lg:items-center lg:justify-between">
          <div class="max-w-2xl">
            <p class="reveal-fade-up text-xs font-semibold uppercase tracking-[0.18em] text-[#77767b]" style="--i: 0">{{ t('home.design.docs.eyebrow') }}</p>
            <h2 class="reveal-fade-up mt-3 text-3xl font-semibold tracking-[-0.05em] text-[#030304] sm:text-4xl" style="--i: 1">{{ t('home.design.docs.title') }}</h2>
            <p class="reveal-fade-up mt-4 text-sm leading-6 text-[#77767b]" style="--i: 2">{{ t('home.design.docs.description') }}</p>
          </div>
          <a data-testid="home-doc-link" :href="docHref" :target="docTarget" :rel="docRel" class="reveal-pop inline-flex shrink-0 items-center justify-center rounded-full bg-[#030304] px-5 py-3 text-sm font-semibold text-white transition-colors hover:bg-[#303034]" style="--i: 2">
            {{ t('home.design.docs.action') }} <span aria-hidden="true" class="ml-2">→</span>
          </a>
        </div>
      </section>

      <section class="reveal-group bg-[#030304] px-4 py-24 text-center text-white sm:px-6 lg:px-8 lg:py-32">
        <div class="mx-auto max-w-4xl">
          <h2 data-testid="home-cta-title" class="reveal-fade-up text-4xl font-semibold tracking-[-0.05em] sm:text-5xl lg:text-6xl" style="--i: 0">{{ t('home.design.cta.title') }}</h2>
          <p class="reveal-fade-up mx-auto mt-5 max-w-2xl text-sm leading-6 text-[#a5a5ae] sm:text-base" style="--i: 1">{{ t('home.design.cta.description') }}</p>
          <div class="reveal-pop mt-9 flex flex-wrap justify-center gap-3" style="--i: 2">
            <router-link :to="primaryDestination" class="inline-flex items-center justify-center rounded-full bg-white px-6 py-3.5 text-sm font-semibold text-[#030304] transition-colors hover:bg-[#f1f0f4]">{{ t('home.design.cta.primary') }}</router-link>
            <a :href="docHref" :target="docTarget" :rel="docRel" class="inline-flex items-center justify-center rounded-full border border-white/15 bg-white/[0.06] px-6 py-3.5 text-sm font-semibold text-white transition-colors hover:bg-white/[0.12]">{{ t('home.design.cta.docs') }}</a>
          </div>
        </div>
      </section>
    </main>

    <footer class="reveal-group border-t border-[#eeedf3] bg-white px-4 py-8 sm:px-6 lg:px-8">
      <div class="mx-auto flex max-w-7xl flex-col gap-5 text-center sm:flex-row sm:items-center sm:justify-between sm:text-left">
        <p class="reveal-fade-up text-xs leading-5 text-[#99989e]" style="--i: 0">© {{ currentYear }} {{ siteName }} · {{ t('home.footer.allRightsReserved') }}</p>
        <div class="reveal-fade-up flex flex-wrap items-center justify-center gap-x-5 gap-y-2 text-xs font-medium text-[#77767b] sm:justify-end" style="--i: 1">
          <a href="#telemetry" class="hover:text-[#030304]">{{ t('home.design.footer.status') }}</a>
          <router-link to="/legal/privacy" class="hover:text-[#030304]">{{ t('home.design.footer.privacy') }}</router-link>
          <router-link to="/legal/terms" class="hover:text-[#030304]">{{ t('home.design.footer.terms') }}</router-link>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="hover:text-[#030304]">{{ t('home.design.footer.github') }}</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { useScrollReveal } from '@/composables/useScrollReveal'

interface Props {
  siteName: string
  siteLogo: string
  siteSubtitle: string
  docUrl: string
  isAuthenticated: boolean
  dashboardPath: string
  showModelPlazaEntry: boolean
}

const props = defineProps<Props>()
const { t } = useI18n()

const githubUrl = 'https://github.com/Wei-Shaw/sub2api'
const currentYear = new Date().getFullYear()

const homeLandingRef = ref<HTMLElement | null>(null)
const heroHaloRef = ref<HTMLElement | null>(null)
const heroTerminalRef = ref<HTMLElement | null>(null)

// Initialize Apple-style scroll reveal observer
useScrollReveal(homeLandingRef)

let targetScrollY = 0
let currentScrollY = 0
let lerpRafId: number | null = null

function isReducedMotion() {
  return (
    typeof window !== 'undefined' &&
    window.matchMedia &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}

function applyParallax(sy: number) {
  if (isReducedMotion()) return
  const progress = Math.min(Math.max(sy / 480, 0), 1)
  if (heroTerminalRef.value) {
    const scale = 0.98 + progress * 0.035
    const rotateX = (1 - progress) * 2.2
    const terminalY = sy * 0.08
    heroTerminalRef.value.style.transform = `translate3d(0, ${terminalY.toFixed(2)}px, 0) scale(${scale.toFixed(4)}) rotateX(${rotateX.toFixed(2)}deg)`
  }
  if (heroHaloRef.value) {
    const haloY = sy * 0.22
    heroHaloRef.value.style.transform = `translate3d(-50%, ${haloY.toFixed(2)}px, 0)`
  }
}

function updateParallax() {
  if (typeof window === 'undefined') return
  const diff = targetScrollY - currentScrollY
  if (Math.abs(diff) > 0.05) {
    currentScrollY += diff * 0.1
    applyParallax(currentScrollY)
    lerpRafId = requestAnimationFrame(updateParallax)
  } else {
    currentScrollY = targetScrollY
    applyParallax(currentScrollY)
    lerpRafId = null
  }
}

function handleScroll() {
  if (isReducedMotion()) return
  targetScrollY = typeof window !== 'undefined' ? window.scrollY : 0
  if (lerpRafId === null) {
    lerpRafId = requestAnimationFrame(updateParallax)
  }
}

const streamingFullText = computed(() => t('home.design.hero.terminalStream'))
const displayedStreamingText = ref(t('home.design.hero.terminalStream'))

let typeInterval: ReturnType<typeof setInterval> | null = null
let typeTimeout: ReturnType<typeof setTimeout> | null = null

function runStreamingTyping() {
  const fullText = streamingFullText.value
  displayedStreamingText.value = ''
  let charIndex = 0
  typeInterval = setInterval(() => {
    if (charIndex < fullText.length) {
      displayedStreamingText.value += fullText.charAt(charIndex)
      charIndex++
    } else {
      if (typeInterval) {
        clearInterval(typeInterval)
        typeInterval = null
      }
      typeTimeout = setTimeout(runStreamingTyping, 6500)
    }
  }, 45)
}

watch(streamingFullText, (newVal) => {
  displayedStreamingText.value = newVal
})

const activeIndex = ref(2)
const windowWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1200)
const trackEl = ref<HTMLElement | null>(null)

function updateWidth() {
  if (typeof window !== 'undefined') {
    windowWidth.value = window.innerWidth
  }
}

onMounted(() => {
  updateWidth()
  window.addEventListener('resize', updateWidth, { passive: true })
  window.addEventListener('scroll', handleScroll, { passive: true })

  if (!isReducedMotion()) {
    targetScrollY = typeof window !== 'undefined' ? window.scrollY : 0
    currentScrollY = targetScrollY
    applyParallax(currentScrollY)
  }

  if (
    typeof window !== 'undefined' &&
    window.matchMedia &&
    window.matchMedia('(prefers-reduced-motion: no-preference)').matches
  ) {
    typeTimeout = setTimeout(runStreamingTyping, 1200)
  }
})

onBeforeUnmount(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', updateWidth)
    window.removeEventListener('scroll', handleScroll)
    if (lerpRafId !== null) {
      cancelAnimationFrame(lerpRafId)
      lerpRafId = null
    }
  }
  if (typeInterval) {
    clearInterval(typeInterval)
    typeInterval = null
  }
  if (typeTimeout) {
    clearTimeout(typeTimeout)
    typeTimeout = null
  }
})

function goTo(index: number) {
  if (index < 0 || index >= modelCards.value.length) return
  activeIndex.value = index
}

function handleCardClick(index: number, event: MouseEvent) {
  if (index === activeIndex.value) return
  if ((event.target as HTMLElement)?.closest('a')) {
    event.preventDefault()
  }
  goTo(index)
}

function carouselXOffset() {
  const isMobile = windowWidth.value < 640
  const isTablet = windowWidth.value >= 640 && windowWidth.value < 1024
  return isMobile ? 220 : isTablet ? 270 : 320
}

// The center card's 3D projection overlaps large parts of its neighbours, so a
// click meant for a side card can land on the center card's surface (a dead
// zone). Map the click x-position to the owning card zone and reroute it so
// every card stays clickable.
function cardZoneIndex(clientX: number) {
  const track = trackEl.value
  if (!track) return activeIndex.value
  const rect = track.getBoundingClientRect()
  if (rect.width <= 0) return activeIndex.value
  const relX = clientX - rect.left - rect.width / 2
  const index = Math.round(relX / carouselXOffset()) + activeIndex.value
  return Math.min(modelCards.value.length - 1, Math.max(0, index))
}

function handleTrackClick(event: MouseEvent) {
  const target = event.target as HTMLElement | null
  const anchor = target?.closest('a')
  if (anchor?.closest('[data-index]')?.getAttribute('data-index') === String(activeIndex.value)) {
    return
  }
  const index = cardZoneIndex(event.clientX)
  if (index === activeIndex.value) return
  if (anchor) {
    event.preventDefault()
  }
  goTo(index)
}

let touchStartX = 0
let touchEndX = 0

function handleTouchStart(e: TouchEvent) {
  if (e.changedTouches.length > 0) {
    touchStartX = e.changedTouches[0].screenX
  }
}

function handleTouchEnd(e: TouchEvent) {
  if (e.changedTouches.length > 0) {
    touchEndX = e.changedTouches[0].screenX
    const deltaX = touchStartX - touchEndX
    if (deltaX > 45) {
      goTo(activeIndex.value + 1)
    } else if (deltaX < -45) {
      goTo(activeIndex.value - 1)
    }
  }
}

function getCardStyle(index: number) {
  const xOffset = carouselXOffset()

  const diff = index - activeIndex.value
  const absDiff = Math.abs(diff)

  if (diff === 0) {
    return {
      transform: 'translateX(-50%) translateZ(60px) scale(1.05) rotateY(0deg)',
      zIndex: 30,
      opacity: 1,
      pointerEvents: 'auto' as const,
      boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.12), 0 0 0 1px rgba(0, 0, 0, 0.06)',
    }
  }

  const factor = absDiff
  const transX = diff * xOffset
  const rotY = diff < 0 ? Math.min(24 + (factor - 1) * 4, 32) : -Math.min(24 + (factor - 1) * 4, 32)
  const transZ = -80 * Math.min(factor, 3)
  const scale = Number(Math.max(0.72, 0.90 - (factor - 1) * 0.08).toFixed(2))
  const opacity = factor > 2 ? 0 : Number(Math.max(0.18, 0.72 - (factor - 1) * 0.35).toFixed(2))
  const zIndex = Math.max(1, 30 - factor)
  const pointerEvents = factor > 2 ? ('none' as const) : ('auto' as const)
  const boxShadow = '0 12px 25px -8px rgba(0, 0, 0, 0.06)'
  const transXStr = transX < 0 ? `calc(-50% - ${Math.abs(transX)}px)` : `calc(-50% + ${transX}px)`

  return {
    transform: `translateX(${transXStr}) translateZ(${transZ}px) scale(${scale}) rotateY(${rotY}deg)`,
    zIndex,
    opacity,
    pointerEvents,
    boxShadow,
  }
}

const primaryDestination = computed(() => (props.isAuthenticated ? props.dashboardPath : '/login'))
const docHref = computed(() => props.docUrl.trim() || '#docs')
const docTarget = computed(() => (props.docUrl.trim() ? '_blank' : undefined))
const docRel = computed(() => (props.docUrl.trim() ? 'noopener noreferrer' : undefined))

const navigationLinks = computed(() => [
  { href: '#models', label: t('home.design.nav.models') },
  { href: '#architecture', label: t('home.design.nav.architecture') },
  { href: '#telemetry', label: t('home.design.nav.telemetry') },
  { href: '#pricing', label: t('home.design.nav.pricing') },
  { href: '#docs', label: t('home.design.nav.docs') },
])

const modelCards = computed(() => [
  {
    key: 'llama',
    mark: 'LL',
    markClass: 'bg-[#6366f1]',
    badgeClass: 'bg-purple-100 text-purple-800',
    badge: t('home.design.models.cards.llama.badge'),
    name: t('home.design.models.cards.llama.name'),
    vendor: t('home.design.models.cards.llama.vendor'),
    description: t('home.design.models.cards.llama.description'),
    note: t('home.design.models.cards.llama.note'),
  },
  {
    key: 'claude',
    mark: 'C',
    markClass: 'bg-[#d97706]',
    badgeClass: 'bg-emerald-100 text-emerald-800',
    badge: t('home.design.models.cards.claude.badge'),
    name: t('home.design.models.cards.claude.name'),
    vendor: t('home.design.models.cards.claude.vendor'),
    description: t('home.design.models.cards.claude.description'),
    note: t('home.design.models.cards.claude.note'),
  },
  {
    key: 'gpt',
    mark: 'G',
    markClass: 'bg-[#10a37f]',
    badgeClass: 'bg-emerald-100 text-emerald-800',
    badge: t('home.design.models.cards.gpt.badge'),
    name: t('home.design.models.cards.gpt.name'),
    vendor: t('home.design.models.cards.gpt.vendor'),
    description: t('home.design.models.cards.gpt.description'),
    note: t('home.design.models.cards.gpt.note'),
  },
  {
    key: 'gemini',
    mark: 'G',
    markClass: 'bg-[#1a73e8]',
    badgeClass: 'bg-emerald-100 text-emerald-800',
    badge: t('home.design.models.cards.gemini.badge'),
    name: t('home.design.models.cards.gemini.name'),
    vendor: t('home.design.models.cards.gemini.vendor'),
    description: t('home.design.models.cards.gemini.description'),
    note: t('home.design.models.cards.gemini.note'),
  },
  {
    key: 'deepseek',
    mark: 'DS',
    markClass: 'bg-[#0e7490]',
    badgeClass: 'bg-blue-100 text-blue-800',
    badge: t('home.design.models.cards.deepseek.badge'),
    name: t('home.design.models.cards.deepseek.name'),
    vendor: t('home.design.models.cards.deepseek.vendor'),
    description: t('home.design.models.cards.deepseek.description'),
    note: t('home.design.models.cards.deepseek.note'),
  },
])

const capabilityCards = computed(() => [
  {
    key: 'access',
    icon: 'bolt' as const,
    title: t('home.design.capabilities.cards.access.title'),
    description: t('home.design.capabilities.cards.access.description'),
    note: t('home.design.capabilities.cards.access.note'),
  },
  {
    key: 'resilience',
    icon: 'shield' as const,
    title: t('home.design.capabilities.cards.resilience.title'),
    description: t('home.design.capabilities.cards.resilience.description'),
    note: t('home.design.capabilities.cards.resilience.note'),
  },
  {
    key: 'billing',
    icon: 'calculator' as const,
    title: t('home.design.capabilities.cards.billing.title'),
    description: t('home.design.capabilities.cards.billing.description'),
    note: t('home.design.capabilities.cards.billing.note'),
  },
])

const telemetryMetrics = computed(() => [
  { key: 'requests', value: t('home.design.telemetry.metrics.requests.value'), label: t('home.design.telemetry.metrics.requests.label') },
  { key: 'availability', value: t('home.design.telemetry.metrics.availability.value'), label: t('home.design.telemetry.metrics.availability.label') },
  { key: 'latency', value: t('home.design.telemetry.metrics.latency.value'), label: t('home.design.telemetry.metrics.latency.label') },
  { key: 'price', value: t('home.design.telemetry.metrics.price.value'), label: t('home.design.telemetry.metrics.price.label') },
])

const pricingItems = computed(() => [
  { key: 'input', value: '0.00', label: t('home.design.pricing.inputLabel') },
  { key: 'output', value: '0.00', label: t('home.design.pricing.outputLabel') },
  { key: 'cache', value: '100%', label: t('home.design.pricing.cacheLabel') },
])
</script>

<style scoped>
:global(html) {
  scroll-behavior: smooth;
}

.home-landing {
  font-family: -apple-system, BlinkMacSystemFont, "SF Pro Display", "SF Pro Text", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
}

.hero-terminal-wrap {
  perspective: 1200px;
}

#heroTerminal {
  transform-origin: center top;
  transition: box-shadow 0.3s ease;
}

.dark-console-bg {
  background: radial-gradient(circle at 10% 10%, #1c1d24 0%, #0d0e12 100%);
}

.stream-cursor {
  display: inline-block;
  width: 7px;
  height: 14px;
  background-color: #34d399;
  vertical-align: middle;
  margin-left: 2px;
  animation: cursorBlink 0.9s infinite step-start;
}

@keyframes cursorBlink {
  50% {
    opacity: 0;
  }
}

.text-balance {
  text-wrap: balance;
}

.carousel-perspective-viewport {
  perspective: 1200px;
  perspective-origin: 50% 50%;
  transform-style: preserve-3d;
}

.coverflow-card {
  position: absolute;
  top: 0;
  left: 50%;
  width: 320px;
  transition: transform 0.6s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.6s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.6s cubic-bezier(0.16, 1, 0.3, 1), border-color 0.4s ease;
  transform-origin: center center;
  user-select: none;
  -webkit-user-select: none;
}

@media (min-width: 640px) {
  .coverflow-card {
    width: 350px;
  }
}

@media (min-width: 1024px) {
  .coverflow-card {
    width: 370px;
  }
}

@media (prefers-reduced-motion: reduce) {
  #heroTerminal,
  .coverflow-card {
    opacity: 1 !important;
    transform: none !important;
    transition: none !important;
  }

  .stream-cursor {
    animation: none !important;
  }
}
</style>
