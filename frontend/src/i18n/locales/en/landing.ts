export default {
  batchImageGuide: {
    title: 'Batch Image Generation',
    description: 'Submit multiple prompts in one job and download the generated images when complete'
  },
  // Home Page
  home: {
    viewOnGithub: 'View on GitHub',
    viewDocs: 'View Documentation',
    docs: 'Docs',
    switchToLight: 'Switch to Light Mode',
    switchToDark: 'Switch to Dark Mode',
    dashboard: 'Dashboard',
    login: 'Login',
    getStarted: 'Get Started',
    goToDashboard: 'Go to Dashboard',
 design: {
  nav: {
   models: 'Model Matrix',
   architecture: 'Core Architecture',
   telemetry: 'Live Telemetry',
   pricing: 'Transparent Billing',
   docs: 'Developer Docs'
  },
  hero: {
   eyebrow: 'Subscription to API Conversion Platform',
   status: 'High-availability protocol aggregation',
   title: 'Unified compute orchestration, <br class="hidden sm:inline">within reach.',
   description: 'Convert multi-source subscriptions into standard OpenAI and Anthropic-compatible APIs. Intelligently route across account clusters with millisecond failover for production AI applications.',
   primary: 'Enter the Console',
   secondary: 'Explore the Model Matrix',
   terminalTitle: 'terminal · anycast-gateway-shanghai',
   terminalProtocol: 'PROTOCOL: HTTP/2 (Multiplexed)',
   terminalActive: 'ACTIVE: 100%',
   terminalRequest: 'POST https://api.sub2.org/v1/chat/completions',
   terminalKey: 'sk-sub2-live-production',
   terminalBody: 'application/json',
   terminalPayload: 'model: \"claude-3-7-sonnet\"',
   terminalRouting: '# Routing to upstream: cluster-us-west-4 (Fallback Pool: Ready)',
   terminalSuccess: '200 OK',
   terminalResult: 'content: \"Hello!\"',
   terminalStream: 'Compute routed in real time. First token latency: 118ms.',
   terminalTelemetry: {
    latency: {
     label: 'Avg TTFT',
     value: '118 ms',
     note: 'Global Anycast Acceleration'
    },
    healing: {
     label: 'Account Self-Healing',
     value: '99.99 %',
     note: 'Zero Rate-Limit Outages'
    },
    overhead: {
     label: 'Protocol Overhead',
     value: '< 1.2 ms',
     note: 'Zero-Copy Engine'
    },
    affinity: {
     label: 'Session Affinity',
     value: '100 %',
     note: 'Context-Aware Routing'
    }
   }
  },
  quickSpecs: {
   automated: 'Fully Automated Subscription-to-API',
   context: 'Native Session Context Preservation',
   billing: 'Accurate Metered Billing & Quota Control',
   sdk: 'OpenAI & Anthropic Official SDK Plug-and-Play'
  },
   models: {
    eyebrow: 'The Lineup',
    title: 'Explore the full model matrix.',
    description: 'One universal API key seamlessly orchestrates flagship foundation models and frontier reasoning systems.',
    official: 'Official support',
    guide: 'Integration guide >',
    cards: {
     llama: {
      badge: 'Open-source cluster',
      name: 'Llama 3.3 / Mistral',
      vendor: 'Meta & Mistral · 70B open-source baseline',
      description: 'Lightweight, precise general-purpose models with global Anycast acceleration and stable throughput for high-concurrency workloads.',
      note: 'Low-cost inference'
     },
     claude: {
      badge: 'Official support',
      name: 'Claude 3.7 / 3.5',
      vendor: 'Anthropic · Sonnet / Opus',
      description: 'Deep semantic reasoning and code refactoring with extended thinking and fast 200K-context streaming output.',
      note: 'Streaming from 130ms'
     },
     gpt: {
      badge: 'Official support',
      name: 'GPT-4.5 / o3-mini',
      vendor: 'OpenAI · Reasoning & Flagship',
      description: 'Flagship multimodal analysis and self-verifying math and physics reasoning with Function Calling and native JSON Mode.',
      note: 'High-concurrency queue'
     },
     gemini: {
      badge: 'Official support',
      name: 'Gemini 2.5 Pro / Flash',
      vendor: 'Google DeepMind · 1M+ Context',
      description: 'Million-token context retrieval, native multimodal understanding, fast responses, and efficient operating costs.',
      note: 'Fast first token'
     },
     deepseek: {
      badge: 'Direct cluster access',
      name: 'DeepSeek R1 / V3',
      vendor: 'DeepSeek Architecture · Full spec',
      description: 'Direct 671B full-spec MoE access with visible long reasoning chains and unrestricted low-latency concurrent throughput.',
      note: 'Full-spec architecture'
     }
    }
   },
  capabilities: {
   eyebrow: 'CORE CAPABILITIES',
   title: 'Engineering first. First-principles design.',
   description: 'Replace bloated relay chains with a compute hub built for high-concurrency production workloads.',
   cards: {
    access: {
     title: '01. One-key integration',
     description: 'Use one unified Master Key to access every connected AI model, without repeating enterprise verification or payment setup across providers.',
     note: 'Zero learning curve · 100% protocol compatible'
    },
    resilience: {
     title: '02. Adaptive self-healing',
     description: 'Global health probes and millisecond routing coordinate upstream accounts, silently retry rate limits, and fail over without interrupting requests.',
     note: '99.99% availability · Fewer errors'
    },
    billing: {
     title: '03. Transparent micro-billing',
     description: 'No base membership fee. Charge precisely for prompt and completion tokens with hard quota protection for teams.',
     note: 'No idle waste · Pay for what you use'
    }
   }
  },
  telemetry: {
   eyebrow: 'LIVE SYSTEM METRICS',
   title: 'Global edge cluster telemetry',
   description: 'Anycast-distributed nodes are monitored continuously with global synchronization every 10 seconds.',
   metrics: {
    requests: { value: '14.8M+', label: 'Requests routed today' },
    availability: { value: '99.98%', label: '30-day end-to-end uptime' },
    latency: { value: '42 ms', label: 'Core network latency in APAC' },
    price: { value: '0.00 ¥', label: '公益 starting threshold' }
   }
  },
  pricing: {
   eyebrow: 'TRANSPARENT BILLING',
   title: 'Transparent micro-billing.',
   description: 'Pay for actual input and output usage with no membership fee or idle waste. Keep costs predictable for individuals and teams.',
   inputLabel: 'Input tokens',
   outputLabel: 'Output tokens',
   cacheLabel: 'Cache hit',
   value: 'Usage based'
  },
  docs: {
   eyebrow: 'DEVELOPER DOCS',
   title: 'Standard APIs. Fast integration.',
   description: 'Use familiar OpenAI and Anthropic API formats to move an existing application onto a resilient compute distribution network in minutes.',
   action: 'View the developer guide'
  },
  cta: {
   title: 'Plug into the next generation of compute distribution.',
   description: 'Register in 30 seconds to receive high-availability trial credits and a standard API gateway endpoint.',
   primary: 'Register free and enter the console',
   docs: 'View the developer guide'
  },
  footer: {
   status: 'System status',
   privacy: 'Privacy',
   terms: 'Terms of service',
   github: 'GitHub repository'
  }
 },
    // User-focused value proposition
    heroSubtitle: 'One Key, All AI Models',
    heroDescription: 'No need to manage multiple subscriptions. Access Claude, GPT, Gemini and more with a single API key',
    tags: {
      subscriptionToApi: 'Subscription to API',
      stickySession: 'Session Persistence',
      realtimeBilling: 'Pay As You Go'
    },
    // Pain points section
    painPoints: {
      title: 'Sound Familiar?',
      items: {
        expensive: {
          title: 'High Subscription Costs',
          desc: 'Paying for multiple AI subscriptions that add up every month'
        },
        complex: {
          title: 'Account Chaos',
          desc: 'Managing scattered accounts and API keys across different platforms'
        },
        unstable: {
          title: 'Service Interruptions',
          desc: 'Single accounts hitting rate limits and disrupting your workflow'
        },
        noControl: {
          title: 'No Usage Control',
          desc: "Can't track where your money goes or limit team member usage"
        }
      }
    },
    // Solutions section
    solutions: {
      title: 'We Solve These Problems',
      subtitle: 'Three simple steps to stress-free AI access'
    },
    features: {
      unifiedGateway: 'One-Click Access',
      unifiedGatewayDesc: 'Get a single API key to call all connected AI models. No separate applications needed.',
      multiAccount: 'Always Reliable',
      multiAccountDesc: 'Smart routing across multiple upstream accounts with automatic failover. Say goodbye to errors.',
      balanceQuota: 'Pay What You Use',
      balanceQuotaDesc: 'Usage-based billing with quota limits. Full visibility into team consumption.'
    },
    // Comparison section
    comparison: {
      title: 'Why Choose Us?',
      headers: {
        feature: 'Comparison',
        official: 'Official Subscriptions',
        us: 'Our Platform'
      },
      items: {
        pricing: {
          feature: 'Pricing',
          official: 'Fixed monthly fee, pay even if unused',
          us: 'Pay only for what you use'
        },
        models: {
          feature: 'Model Selection',
          official: 'Single provider only',
          us: 'Switch between models freely'
        },
        management: {
          feature: 'Account Management',
          official: 'Manage each service separately',
          us: 'Unified key, one dashboard'
        },
        stability: {
          feature: 'Stability',
          official: 'Single account rate limits',
          us: 'Multi-account pool, auto-failover'
        },
        control: {
          feature: 'Usage Control',
          official: 'Not available',
          us: 'Quotas & detailed analytics'
        }
      }
    },
    providers: {
      title: 'Supported AI Models',
      description: 'One API, Multiple Choices',
      supported: 'Supported',
      soon: 'Soon',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: 'More'
    },
    // CTA section
    cta: {
      title: 'Ready to Get Started?',
      description: 'Sign up now and get free trial credits to experience seamless AI access',
      button: 'Sign Up Free'
    },
    footer: {
      allRightsReserved: 'All rights reserved.'
    }
  },

  // Key Usage Query Page
  keyUsage: {
    title: 'API Key Usage',
    subtitle: 'Enter your API Key to view real-time spending and usage status',
    placeholder: 'sk-ant-mirror-xxxxxxxxxxxx',
    query: 'Query',
    querying: 'Querying...',
    privacyNote: 'Your Key is processed locally in the browser and will not be stored',
    dateRange: 'Date Range:',
    dateRangeToday: 'Today',
    dateRange7d: '7 Days',
    dateRange30d: '30 Days',
    dateRange90d: '90 Days',
    dateRangeCustom: 'Custom',
    apply: 'Apply',
    used: 'Used',
    detailInfo: 'Detail Information',
    tokenStats: 'Token Statistics',
    dailyDetail: 'Daily Detail',
    modelStats: 'Model Usage Statistics',
    // Table headers
    date: 'Date',
    model: 'Model',
    requests: 'Requests',
    inputTokens: 'Input Tokens',
    outputTokens: 'Output Tokens',
    cacheCreationTokens: 'Cache Creation',
    cacheReadTokens: 'Cache Read',
    cacheWriteTokens: 'Cache Write',
    totalTokens: 'Total Tokens',
    cost: 'Cost',
    // Status
    quotaMode: 'Key Quota Mode',
    walletBalance: 'Wallet Balance',
    // Ring card titles
    totalQuota: 'Total Quota',
    limit5h: '5-Hour Limit',
    limitDaily: 'Daily Limit',
    limit7d: '7-Day Limit',
    limitWeekly: 'Weekly Limit',
    limitMonthly: 'Monthly Limit',
    // Detail rows
    remainingQuota: 'Remaining Quota',
    expiresAt: 'Expires At',
    todayExpires: '(expires today)',
    daysLeft: '({days} days)',
    usedQuota: 'Used Quota',
    resetNow: 'Resetting soon',
    subscriptionType: 'Subscription Type',
    billingType: 'Billing Type',
    subscriptionExpires: 'Subscription Expires',
    // Usage stat cells
    todayRequests: 'Today Requests',
    todayInputTokens: 'Today Input',
    todayOutputTokens: 'Today Output',
    todayTokens: 'Today Tokens',
    todayCacheCreation: 'Today Cache Creation',
    todayCacheRead: 'Today Cache Read',
    todayCost: 'Today Cost',
    rpmTpm: 'RPM / TPM',
    totalRequests: 'Total Requests',
    totalInputTokens: 'Total Input',
    totalOutputTokens: 'Total Output',
    totalTokensLabel: 'Total Tokens',
    totalCacheCreation: 'Total Cache Creation',
    totalCacheRead: 'Total Cache Read',
    totalCost: 'Total Cost',
    avgDuration: 'Avg Duration',
    // Messages
    enterApiKey: 'Please enter an API Key',
    querySuccess: 'Query successful',
    queryFailed: 'Query failed',
    queryFailedRetry: 'Query failed, please try again later',
    noDailyUsage: 'No daily usage data',
  },

  // Setup Wizard
  setup: {
    title: 'Sub2API Setup',
    description: 'Configure your Sub2API instance',
    database: {
      title: 'Database Configuration',
      description: 'Connect to your PostgreSQL database',
      host: 'Host',
      port: 'Port',
      username: 'Username',
      password: 'Password',
      databaseName: 'Database Name',
      sslMode: 'SSL Mode',
      passwordPlaceholder: 'Password',
      ssl: {
        disable: 'Disable',
        require: 'Require',
        verifyCa: 'Verify CA',
        verifyFull: 'Verify Full'
      }
    },
    redis: {
      title: 'Redis Configuration',
      description: 'Connect to your Redis server',
      host: 'Host',
      port: 'Port',
      username: 'Username (optional)',
      password: 'Password (optional)',
      database: 'Database',
      usernamePlaceholder: 'Leave empty for default user',
      passwordPlaceholder: 'Password',
      enableTls: 'Enable TLS',
      enableTlsHint: 'Use TLS when connecting to Redis (public CA certs)'
    },
    admin: {
      title: 'Admin Account',
      description: 'Create your administrator account',
      email: 'Email',
      password: 'Password',
      confirmPassword: 'Confirm Password',
      passwordPlaceholder: 'Min 8 characters',
      confirmPasswordPlaceholder: 'Confirm password',
      passwordMismatch: 'Passwords do not match'
    },
    ready: {
      title: 'Ready to Install',
      description: 'Review your configuration and complete setup',
      database: 'Database',
      redis: 'Redis',
      adminEmail: 'Admin Email'
    },
    status: {
      testing: 'Testing...',
      success: 'Connection Successful',
      testConnection: 'Test Connection',
      installing: 'Installing...',
      completeInstallation: 'Complete Installation',
      completed: 'Installation completed!',
      redirecting: 'Redirecting to login page...',
      restarting: 'Service is restarting, please wait...',
      timeout: 'Service restart is taking longer than expected. Please refresh the page manually.'
    }
  },

  // Common
}
