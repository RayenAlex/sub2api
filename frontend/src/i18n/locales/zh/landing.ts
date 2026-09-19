export default {
  batchImageGuide: {
    title: '图片批量生成',
    description: '一次提交多条提示词，任务完成后可统一下载图片结果'
  },
  // Home Page
  home: {
    viewOnGithub: '在 GitHub 上查看',
    viewDocs: '查看文档',
    docs: '文档',
    switchToLight: '切换到浅色模式',
    switchToDark: '切换到深色模式',
    dashboard: '控制台',
    login: '登录',
    getStarted: '立即开始',
    goToDashboard: '进入控制台',
 design: {
  nav: {
   models: '模型矩阵',
   architecture: '核心架构',
   telemetry: '实时遥测',
   pricing: '透明计费',
   docs: '接入文档'
  },
  hero: {
   eyebrow: 'Subscription to API Conversion Platform',
   status: '全协议高可用聚合',
   title: '算力统一调度，<br class="hidden sm:inline">如电力般触手可及。',
   description: '一键将多源订阅转化为标准 OpenAI & Anthropic 格式 API。智能调度集群账号，毫秒级故障旁路切换，为生产级 AI 原生应用而生。',
   primary: '立即进入控制台',
   secondary: '探索模型矩阵',
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
   terminalStream: '算力已实时调度完成。首字响应延迟 118ms。',
   terminalTelemetry: {
    latency: {
     label: '平均首字延迟',
     value: '118 ms',
     note: '全球 Anycast 加速'
    },
    healing: {
     label: '多账号自愈成功率',
     value: '99.99 %',
     note: '0 频发频控中断'
    },
    overhead: {
     label: '协议转换损耗',
     value: '< 1.2 ms',
     note: '零内存拷贝引擎'
    },
    affinity: {
     label: '会话粘性保持',
     value: '100 %',
     note: '智能上下文感知'
    }
   }
  },
  quickSpecs: {
   automated: '订阅转 API 全自动化',
   context: '原生会话上下文保持',
   billing: '精确按量计费与配额控制',
   sdk: 'OpenAI & Anthropic 官方 SDK 即插即用'
  },
   models: {
    eyebrow: 'The Lineup',
    title: '探索全系模型矩阵。',
    description: '一个通用 API 密钥，无缝调度顶尖主流基础大模型与前沿推理体系。',
    official: '官方全量支持',
    guide: '接入指南 >',
    cards: {
     llama: {
      badge: '高并发开源集群',
      name: 'Llama 3.3 / Mistral',
      vendor: 'Meta & Mistral · 70B 开源基准',
      description: '极度轻量且精准的泛化开源生态支撑，全球多机房 Anycast 加速部署，高并发海量任务稳定吞吐。',
      note: '低成本极速并发'
     },
     claude: {
      badge: '官方全量支持',
      name: 'Claude 3.7 / 3.5',
      vendor: 'Anthropic · Sonnet / Opus',
      description: '极致的代码逻辑重构与深层语义推理。支持超长思考链 (Extended Thinking) 与 200K 上下文极速输出。',
      note: '流式低至 130ms'
     },
     gpt: {
      badge: '官方全量支持',
      name: 'GPT-4.5 / o3-mini',
      vendor: 'OpenAI · Reasoning & Flagship',
      description: '旗舰级多模态分析、高难数学物理自验证，兼容 Function Calling 与原生 JSON Mode 严谨约束。',
      note: '高并发并发队列'
     },
     gemini: {
      badge: '官方全量支持',
      name: 'Gemini 2.5 Pro / Flash',
      vendor: 'Google DeepMind · 1M+ Context',
      description: '百万级超长上下文窗口无衰减检索，全模态音画原生理解，超快生成响应与极限成本优势。',
      note: '首字响应极速'
     },
     deepseek: {
      badge: '集群直通接入',
      name: 'DeepSeek R1 / V3',
      vendor: 'DeepSeek Architecture · 满血版',
      description: '671B 满血 MoE 架构直连，深度长思维链思维展示。低时延无限制并发吞吐调度。',
      note: '满血架构无阉割'
     }
    }
   },
  capabilities: {
   eyebrow: 'CORE CAPABILITIES',
   title: '工程至上。第一性原理设计。',
   description: '摒弃传统中转平台的臃肿链路，打造专为高并发生产环境定制的算力枢纽。',
   cards: {
    access: {
     title: '01. 极简一键接入',
     description: '仅需一个统一 Master Key，即可调配全量已接入的全球 AI 模型。完全免除在十余家模型供应商间重复申请企业实名与复杂的信用卡绑卡流程。',
     note: '零学习成本 · 100% 官方协议兼容'
    },
    resilience: {
     title: '02. 自适应弹性自愈',
     description: '自研全局健康度感应探针。毫秒级智能调度多个上游发放账号，自动在速率限制出现时完成零感知静默重试与无缝故障旁路切换。',
     note: '99.99% 可用性 · 告别频繁报错'
    },
    billing: {
     title: '03. 透明微量计费',
     description: '完全摒弃保底会员月租机制。按实际流式生成的 Prompt 与 Completion Token 精确扣费，支持企业级多成员硬性配额上限保护。',
     note: '无任何闲置损耗 · 用多少付多少'
    }
   }
  },
  telemetry: {
   eyebrow: '实时系统运行指标',
   title: '全球边缘集群遥测状态',
   description: '基于 Anycast 分布式节点全天候监控，数据每 10 秒全局同步刷新。',
   metrics: {
    requests: { value: '14.8M+', label: '今日完成调度请求 (REQUESTS)' },
    availability: { value: '99.98%', label: '30 天全链路可用率' },
    latency: { value: '42 ms', label: '亚太核心网内延迟' },
    price: { value: '0.00 ¥', label: '公益起步费用门槛' }
   }
  },
  pricing: {
   eyebrow: 'TRANSPARENT BILLING',
   title: '透明的微量计费。',
   description: '按实际输入与输出用量计费，没有月租、没有闲置损耗，为个人开发者和团队提供可预期的成本控制。',
   inputLabel: '输入 Token',
   outputLabel: '输出 Token',
   cacheLabel: '缓存命中',
   value: '按量计费'
  },
  docs: {
   eyebrow: 'DEVELOPER DOCS',
   title: '标准 API，快速接入。',
   description: '使用熟悉的 OpenAI 与 Anthropic API 格式，在几分钟内将现有应用切换到高可用算力分发网络。',
   action: '查看开发者接入指南'
  },
  cta: {
   title: '立即接入下一代算力分发网络。',
   description: '只需 30 秒注册，立即获取高可用测试额度与标准 API 终端网关地址。',
   primary: '免费注册并进入控制台',
   docs: '查看开发者接入指南'
  },
  footer: {
   status: '系统状态',
   privacy: '隐私声明',
   terms: '服务条款',
   github: 'GitHub 仓库'
  }
 },
    // 新增：面向用户的价值主张
    heroSubtitle: '一个密钥，畅用多个 AI 模型',
    heroDescription: '无需管理多个订阅账号，一站式接入 Claude、GPT、Gemini 等主流 AI 服务',
    tags: {
      subscriptionToApi: '订阅转 API',
      stickySession: '会话保持',
      realtimeBilling: '按量计费'
    },
    // 用户痛点区块
    painPoints: {
      title: '你是否也遇到这些问题？',
      items: {
        expensive: {
          title: '订阅费用高',
          desc: '每个 AI 服务都要单独订阅，每月支出越来越多'
        },
        complex: {
          title: '多账号难管理',
          desc: '不同平台的账号、密钥分散各处，管理起来很麻烦'
        },
        unstable: {
          title: '服务不稳定',
          desc: '单一账号容易触发限制，影响正常使用'
        },
        noControl: {
          title: '用量无法控制',
          desc: '不知道钱花在哪了，也无法限制团队成员的使用'
        }
      }
    },
    // 解决方案区块
    solutions: {
      title: '我们帮你解决',
      subtitle: '简单三步，开始省心使用 AI'
    },
    features: {
      unifiedGateway: '一键接入',
      unifiedGatewayDesc: '获取一个 API 密钥，即可调用所有已接入的 AI 模型，无需分别申请。',
      multiAccount: '稳定可靠',
      multiAccountDesc: '智能调度多个上游账号，自动切换和负载均衡，告别频繁报错。',
      balanceQuota: '用多少付多少',
      balanceQuotaDesc: '按实际使用量计费，支持设置配额上限，团队用量一目了然。'
    },
    // 优势对比
    comparison: {
      title: '为什么选择我们？',
      headers: {
        feature: '对比项',
        official: '官方订阅',
        us: '本平台'
      },
      items: {
        pricing: {
          feature: '付费方式',
          official: '固定月费，用不完也付',
          us: '按量付费，用多少付多少'
        },
        models: {
          feature: '模型选择',
          official: '单一服务商',
          us: '多模型随意切换'
        },
        management: {
          feature: '账号管理',
          official: '每个服务单独管理',
          us: '统一密钥，一站管理'
        },
        stability: {
          feature: '服务稳定性',
          official: '单账号易触发限制',
          us: '多账号池，自动切换'
        },
        control: {
          feature: '用量控制',
          official: '无法限制',
          us: '可设配额、查明细'
        }
      }
    },
    providers: {
      title: '已支持的 AI 模型',
      description: '一个 API，多种选择',
      supported: '已支持',
      soon: '即将推出',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: '更多'
    },
    // CTA 区块
    cta: {
      title: '准备好开始了吗？',
      description: '注册即可获得免费试用额度，体验一站式 AI 服务',
      button: '免费注册'
    },
    footer: {
      allRightsReserved: '保留所有权利。'
    }
  },

  // Key Usage Query Page
  keyUsage: {
    title: 'API Key 用量查询',
    subtitle: '输入您的 API Key 以查看实时消费金额与使用状态',
    placeholder: 'sk-ant-mirror-xxxxxxxxxxxx',
    query: '查询',
    querying: '查询中...',
    privacyNote: '您的 Key 仅在浏览器本地处理，不会被存储',
    dateRange: '统计范围:',
    dateRangeToday: '今日',
    dateRange7d: '7 天',
    dateRange30d: '30 天',
    dateRange90d: '90 天',
    dateRangeCustom: '自定义',
    apply: '应用',
    used: '已使用',
    detailInfo: '详细信息',
    tokenStats: 'Token 统计',
    dailyDetail: '按日明细',
    modelStats: '模型用量统计',
    // Table headers
    date: '日期',
    model: '模型',
    requests: '请求数',
    inputTokens: '输入 Tokens',
    outputTokens: '输出 Tokens',
    cacheCreationTokens: '缓存创建',
    cacheReadTokens: '缓存读取',
    cacheWriteTokens: '缓存写入',
    totalTokens: '总 Tokens',
    cost: '费用',
    // Status
    quotaMode: 'Key 限额模式',
    walletBalance: '钱包余额',
    // Ring card titles
    totalQuota: '总额度',
    limit5h: '5 小时限额',
    limitDaily: '日限额',
    limit7d: '7 天限额',
    limitWeekly: '周限额',
    limitMonthly: '月限额',
    // Detail rows
    remainingQuota: '剩余额度',
    expiresAt: '过期时间',
    todayExpires: '(今日到期)',
    daysLeft: '({days} 天)',
    usedQuota: '已用额度',
    resetNow: '即将重置',
    subscriptionType: '订阅类型',
    billingType: '计费方式',
    subscriptionExpires: '订阅到期',
    // Usage stat cells
    todayRequests: '今日请求',
    todayInputTokens: '今日输入',
    todayOutputTokens: '今日输出',
    todayTokens: '今日 Tokens',
    todayCacheCreation: '今日缓存创建',
    todayCacheRead: '今日缓存读取',
    todayCost: '今日费用',
    rpmTpm: 'RPM / TPM',
    totalRequests: '累计请求',
    totalInputTokens: '累计输入',
    totalOutputTokens: '累计输出',
    totalTokensLabel: '累计 Tokens',
    totalCacheCreation: '累计缓存创建',
    totalCacheRead: '累计缓存读取',
    totalCost: '累计费用',
    avgDuration: '平均耗时',
    // Messages
    enterApiKey: '请输入 API Key',
    querySuccess: '查询成功',
    queryFailed: '查询失败',
    queryFailedRetry: '查询失败，请稍后重试',
    noDailyUsage: '暂无按日用量数据',
  },

  // Setup Wizard
  setup: {
    title: 'Sub2API 安装向导',
    description: '配置您的 Sub2API 实例',
    database: {
      title: '数据库配置',
      description: '连接到您的 PostgreSQL 数据库',
      host: '主机',
      port: '端口',
      username: '用户名',
      password: '密码',
      databaseName: '数据库名称',
      sslMode: 'SSL 模式',
      passwordPlaceholder: '密码',
      ssl: {
        disable: '禁用',
        require: '要求',
        verifyCa: '验证 CA',
        verifyFull: '完全验证'
      }
    },
    redis: {
      title: 'Redis 配置',
      description: '连接到您的 Redis 服务器',
      host: '主机',
      port: '端口',
      username: '用户名（可选）',
      password: '密码（可选）',
      database: '数据库',
      usernamePlaceholder: '默认用户留空',
      passwordPlaceholder: '密码',
      enableTls: '启用 TLS',
      enableTlsHint: '连接 Redis 时使用 TLS（公共 CA 证书）'
    },
    admin: {
      title: '管理员账户',
      description: '创建您的管理员账户',
      email: '邮箱',
      password: '密码',
      confirmPassword: '确认密码',
      passwordPlaceholder: '至少 8 个字符',
      confirmPasswordPlaceholder: '确认密码',
      passwordMismatch: '密码不匹配'
    },
    ready: {
      title: '准备安装',
      description: '检查您的配置并完成安装',
      database: '数据库',
      redis: 'Redis',
      adminEmail: '管理员邮箱'
    },
    status: {
      testing: '测试中...',
      success: '连接成功',
      testConnection: '测试连接',
      installing: '安装中...',
      completeInstallation: '完成安装',
      completed: '安装完成！',
      redirecting: '正在跳转到登录页面...',
      restarting: '服务正在重启，请稍候...',
      timeout: '服务重启时间超出预期，请手动刷新页面。'
    }
  },

  // Common
}
