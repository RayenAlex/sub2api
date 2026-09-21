# Rayen 提交 Code Review — 以 dev 为底的落地页重构 (25f44ece)

> **审查分支**
> - 基准分支: `dev` → `876ddb77 feat: unify usage telemetry interface` (Rayen 重构前的最近 dev)
> - 审查分支: `review/rayen-from-dev` (基于 `dev` = `2b105685`) 与 `review/base-before-rayen` (876ddb77) + `review/rayen-compare` (876ddb77 + 25f44ece cherry-pick)
> - 目标提交: `25f44ecef feat: redesign landing page with 3D model carousel and Apple-style scroll reveal` — Author: rayen <rayenofjri@gmail.com>
> - 变更文件 9 个，+1814 / -571

```
frontend/src/components/home/HomeLandingPage.vue         | 900 +++
frontend/src/components/home/__tests__/HomeLandingPage.spec.ts | 209 +++
frontend/src/composables/__tests__/useScrollReveal.spec.ts | 164 +++
frontend/src/composables/useScrollReveal.ts               | 154 +++
frontend/src/i18n/locales/en/landing.ts                   | 160 +++
frontend/src/i18n/locales/zh/landing.ts                   | 160 +++
frontend/src/style.css                                     |  56 ++
frontend/src/views/HomeView.vue                            | 580 +------------
frontend/src/views/__tests__/HomeView.compact.spec.ts      |   2 +-
```

---

## 1. 总体评价

**方向正确，工程化意识强：**

- 把原来 600+ 行的 `HomeView.vue` 拆成 `HomeLandingPage.vue` + 保留 `compact/homeContent` 兼容分支，职责更清晰，符合 `DEV_GUIDE` 中“按平台分组”的思想。
- 新增 `useScrollReveal` composable 抽离 `IntersectionObserver` 逻辑，支持 `prefers-reduced-motion`、SSR guard、`.reveal-group` 协调时序，考虑了无障碍。
- 3D Coverflow 轮播实现完整：center-focused、side dimming、dots/arrows、touch swipe、keyboard nav、click-zone 重路由（解决中心卡片遮挡侧卡片的死区问题），细节到位。
- i18n 完整补齐 `zh/en landing.ts`，测试覆盖 `HomeLandingPage.spec.ts` + `useScrollReveal.spec.ts`，符合项目“前端改动必须带测试”的要求。
- 提交信息规范，列出设计点。

**风险等级：中低** — 纯前端展示层，不涉及计费、鉴权、调度等核心链路，但仍有若干性能、可访问性、样式隔离问题需要修。

---

## 2. 逐文件精读

### 2.1 `HomeLandingPage.vue` (900 行)

**优点：**
- 模板结构按 Apple 官网分段：`#hero #models #architecture #telemetry #pricing #docs`，锚点导航 `navigationLinks` computed。
- Hero 区用 `perspective: 1200px` + `heroTerminalRef` 视差，`streamingText` 打字机效果，`heroHaloRef` 光晕，视觉层次好。
- Coverflow 实现：
  ```ts
  function carouselXOffset() { return isMobile ? 220 : isTablet ? 270 : 320 }
  function cardZoneIndex(clientX) { /* 根据 track 居中偏移算 zone */ }
  function handleTrackClick() { /* dead-zone 重路由 */ }
  ```
  解决真实 3D 投影遮挡问题，思路巧妙。
- `goTo` 边界保护，`activeIndex` 初始 2 (GPT)，`prev/next` disabled 状态正确。

**问题：**

**🔴 P1 — 内存泄漏与全局监听未限流**
```ts
onMounted(() => {
  window.addEventListener('resize', updateWidth)
  window.addEventListener('scroll', handleScroll)
  lerpRafId = requestAnimationFrame(lerpLoop)
})
```
- `resize` / `scroll` 未做 `throttle`/`passive`，滚动时 `handleScroll` 会频繁触发，移动端掉帧。
- `lerpRafId` 的 `lerpLoop` 在 `prefers-reduced-motion: reduce` 时仍可能运行，浪费 CPU。
- 建议：`resize` 用 `useDebounceFn(100)`，`scroll` 加 `{passive:true}` 且 `requestAnimationFrame` 节流；`isReducedMotion()` 时直接跳过 lerp。

**🟠 P2 — `v-html` 潜在 XSS (虽不在本文件，但在 `HomeView.vue` 保留分支)**
```vue
<div v-else v-html="homeContent"></div>
```
注释写“admin-only, XSS risk is acceptable”，但 `homeContent` 来自 `publicSettings.home_content`，若管理员账号被劫持，可直接存储型 XSS。建议即使 admin-only 也走 `DOMPurify.sanitize`。

**🟠 P2 — 硬编码色值与暗色模式缺失**
- `bg-white text-[#1d1d1f]` 写死亮色，未适配 `dark`。原 `HomeView.vue` 有 `dark:bg-dark-950`，新 landing 页强制亮色，切暗色时会闪白。
- 建议：用 CSS 变量或 `dark:` 变体，或至少在 `home-landing` 根加 `color-scheme: light` 声明意图。

**🟠 P2 — 可访问性**
- Coverflow 卡片 `div` 用 `@click`，但无 `role="button"`、`tabindex="0"`、`@keydown.enter`，键盘用户无法激活侧卡。
- `prevModelBtn` / `nextModelBtn` 有 `disabled`，但无 `aria-label`。
- 建议：侧卡加 `role="button" tabindex="0" aria-label="View {{name}}"`。

**🟡 P3 — 性能：900 行单文件**
- `modelCards`、`capabilityCards`、`telemetryMetrics` 都是 computed 返回新数组，每次渲染都会重新创建，触发子组件 diff。
- 建议：抽成 `constants/landing.ts` 静态数组，`t()` 用函数式 i18n 或 `computed` 内 `t` 已可，但数组本身可 `shallowRef`。

**🟡 P3 — 样式隔离**
- `<style scoped>` 中有 `:global(html) { scroll-behavior: smooth; }`，会污染全局。应放到 `style.css` 或 `App.vue`。
- `.coverflow-card { transition: transform 0.6s ... }` 用 `all` 更好指定 `transform, opacity`，避免意外触发 `box-shadow` 重绘。

### 2.2 `useScrollReveal.ts` (154 行)

**优点：**
- SSR guard `typeof window === 'undefined'`，`matchMedia` 存在检查，`try/catch` 包裹 `IntersectionObserver`。
- `reveal-group` 协调子元素，避免 stagger 错乱。
- `scopeRef` 支持局部容器，`.reveal-ready` 防止 FOIC，设计细致。

**问题：**

**🟠 P2 — Observer 未在 scope 变化时重建**
- `scopeRef` 是 Ref，但 `init()` 只在 `onMounted` 调一次，若 `scopeRef.value` 后续变化（例如异步加载），不会重新观察。
- 建议：`watch(scopeRef, () => { cleanup(); nextTick(init) })`。

**🟡 P3 — `document.querySelector('.home-landing')` 全局查询**
- 若页面有多个 `.home-landing`（例如测试环境），会取到错误的。应优先 `scopeRef.value`，否则 `return`，不 fallback 到 `document.body`。
- 当前 `container = scopeRef?.value || document.querySelector('.home-landing') || document.body`，`document.body` 会导致全页所有 `.reveal-*` 都被触发，性能差。

**🟡 P3 — Directive 未注册**
- 导出了 `vReveal`，但 `HomeLandingPage.vue` 未 `app.directive('reveal', vReveal)`，也未在模板中使用，死代码。
- 建议：要么在模板用 `v-reveal:pop="i"`，要么删除导出。

**🟡 P3 — 阈值与 rootMargin 可配置性**
- `rootMargin = '0px 0px -25% 0px'` 注释说 75% 视口触发，但 0.05 threshold 过小，快速滚动可能跳过。建议 threshold 改为 `[0, 0.15]`。

### 2.3 `HomeView.vue` 重构

**优点：**
- 保留 `hasHomeContent` / `compactHomeEnabled` 双兼容分支，`HomeLandingPage` 作为 default，迁移风险小。
- `sanitizeUrl` 对 `siteLogo` / `docUrl` 做净化，延续安全实践。

**问题：**
- `toggleTheme` 直接操作 `document.documentElement.classList`，与 `useAppStore` 的 theme 状态不同步，可能导致 Pinia 与 DOM 不一致。
- `currentYear` 用 `computed(() => new Date().getFullYear())`，每年都会变但无响应式依赖，其实可用常量。

### 2.4 `style.css` 新增 56 行

**优点：**
- `.reveal-ready .reveal-fade-up:not(.is-revealed)` 用 `pointer-events:none` 避免隐藏元素可点击，细节好。
- `prefers-reduced-motion` 三档降级完整。

**问题：**
- `.reveal-fade-up, .reveal-pop { will-change: opacity, transform; }` 全局常驻 `will-change` 会增加内存，应在 `.is-revealed` 后移除，或用 `contain: layout`。
- `transition-delay: calc(var(--i,0)*90ms)` 依赖 `--i` 内联样式，若 `--i` 未设会退 0，但未设上限，`--i=10` 时延迟 900ms，体验差，建议 `clamp(--i,0,6)`。

### 2.5 i18n `landing.ts`

- `en` / `zh` 各 160 行，key 完整，但部分文案硬编码英文品牌名在 zh 文件中（如 `Anycast`、`TTFT`），建议保留英文术语但加括号中文解释，已有部分做到。
- `home.design.hero.title` 含 `<br class="hidden sm:inline">` HTML，`v-html="t(...)"` 会解析 HTML，需确保 i18n 值受控，否则 XSS。当前是静态翻译，可接受，但建议用 `i18n-t` 组件。

### 2.6 测试

**优点：**
- `HomeLandingPage.spec.ts` 覆盖：sections 存在、primary destination、model plaza 条件、doc URL、coverflow 初始 index、next/dot/card 点击、keyboard ArrowRight、dead-zone 重路由、stagger 属性。
- `useScrollReveal.spec.ts` mock `IntersectionObserver`、`matchMedia`，测 reduced-motion 立即 reveal，逻辑完整。

**问题：**
- `getBoundingClientRect` mock 返回固定 1280px，但 `carouselXOffset()` 依赖 `windowWidth` (resize 监听)，测试中 `windowWidth` 未 mock，可能在 CI 中 `window.innerWidth=1024` 导致偏移计算不稳定。
- 未测 `touchStart`/`touchEnd` swipe（阈值 45px），建议补一个 `touch` 用例。
- `HomeView.compact.spec.ts` 只改 2 行，但未测 `homeContent` iframe vs html 分支。

---

## 3. 安全与合规

- **XSS**：`HomeView.vue` 的 `v-html="homeContent"` 是唯一风险点，admin-only 可接受，但建议加 `DOMPurify`。
- **CSP**：`DefaultCSPPolicy` 有 `__CSP_NONCE__`，新 landing 页内联 `<style scoped>` 不含 nonce，若开启严格 CSP 会被拦截。Vite 默认会提取，建议检查 `vite.config` 的 `csp` 选项。
- **No `InsecureSkipVerify`**：本次变更无后端，安全合规。

---

## 4. 性能建议

1. **滚动与 resize 限流**：`handleScroll` + `updateWidth` 用 `requestAnimationFrame` + `debounce`。
2. **图片懒加载**：`siteLogo` 是 `img`，未加 `loading="lazy"`，首屏 LCP 可能受影响。
3. **3D 卡片**：`transform: translateZ(60px)` 会创建新合成层，5 张卡同时 `will-change`，移动端 GPU 内存高。建议仅 `activeIndex` 加 `will-change`，其余移除。
4. **打字机效果**：`setInterval` 打字未用 `requestAnimationFrame`，且 `displayedStreamingText` 每次 `slice` 产生新字符串，频繁触发 Vue 响应式。可用 `shallowRef`。

---

## 5. 可维护性建议

- **拆分**：`HomeLandingPage.vue` 900 行，建议拆 `HeroSection.vue` / `ModelCoverflow.vue` / `CapabilityGrid.vue` / `TelemetrySection.vue`，每个 <200 行。
- **常量抽离**：`modelCards` 的 `markClass`、`badgeClass` 硬编码 Tailwind，建议抽 `constants/landing.ts`。
- **类型**：`modelCards` 未定义 interface，`any` 推断，补 `type ModelCard = {key, mark, ...}`。
- **a11y**：补 `aria-roledescription="carousel"`、`aria-live="polite"`。

---

## 6. 与 dev 基准的集成风险

- `dev` 基准 `876ddb77` 已有 `feat: unify usage telemetry interface`，本次 landing 页新增 `telemetry` 区块文案与 `usage` 模块的 `cache hit rate` 无冲突。
- `style.css` 在 `dev` 中有大量 `telemetry-trend-*` 样式，本次新增的 `reveal-*` 未命名冲突，安全。
- `HomeView.vue` 的 `compactHomeEnabled` 在 `dev` 中已存在，保留兼容分支，线上若开启 `compact_home` 仍走旧 compact，不受新 landing 影响，回滚友好。

---

## 7. 结论与 Action Items

**总体：LGTM with comments，功能完整，测试充分，可合并，但建议修 P1/P2 后再发版。**

**必须修 (P1)：**
- [ ] `resize` / `scroll` 加 `passive` + `throttle`/`rAF`
- [ ] `will-change` 仅 active 卡，移除全局常驻

**建议修 (P2)：**
- [ ] Coverflow 卡片加 `role="button" tabindex` + `aria-label`
- [ ] `scopeRef` 变化时重建 observer，`document.body` fallback 去掉
- [ ] `HomeView.vue` 的 `v-html` 加 `DOMPurify` 或注释说明已做后端消毒
- [ ] `:global(html)` 移出 scoped

**可选 (P3)：**
- [ ] 拆分 900 行组件
- [ ] 补 touch swipe 测试
- [ ] `vReveal` directive 要么使用要么删除

**分支说明：**
- 已创建 `review/rayen-from-dev` (基于 `dev`，含 Rayen 全部提交) 供 CI 跑 `pnpm test`
- `review/base-before-rayen` (876ddb77) 是纯 dev 基准
- `review/rayen-compare` 是 `base + 25f44ece` 的最小可审查单元，`git diff --cached` 即本次 Review 范围

> 后续若需我直接在 `review/rayen-compare` 上提修复 commit，可在 `arena/01a0c4b3-sub2api` 分支上 `git checkout review/rayen-compare` 继续改。

