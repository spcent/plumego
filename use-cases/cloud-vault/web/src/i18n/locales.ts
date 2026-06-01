export interface Translations {
  common: {
    save: string
    cancel: string
    delete: string
    edit: string
    create: string
    search: string
    loading: string
    error: string
    success: string
    confirm: string
  }
  nav: {
    vault: string
    search: string
    import: string
    index: string
    duplicates: string
    collections: string
    topics: string
    review: string
    aiTasks: string
    prompts: string
    system: string
    settings: string
    account: string
    security: string
    logout: string
    login: string
  }
  auth: {
    signIn: string
    signOut: string
    username: string
    password: string
    currentPassword: string
    newPassword: string
    confirmPassword: string
    changePassword: string
    loginFailed: string
    invalidCredentials: string
    rateLimited: string
    updateProfile: string
    displayName: string
    email: string
    locale: string
    theme: string
    passwordMismatch: string
    passwordTooShort: string
    invalidCurrentPassword: string
    signingIn: string
  }
  setup: {
    title: string
    description: string
    createAccount: string
    creating: string
    setupFailed: string
    alreadyInitialized: string
  }
  pages: {
    vault: {
      title: string
      documents: string
      noDocuments: string
      uploadDocument: string
    }
    search: {
      title: string
      placeholder: string
      noResults: string
    }
    import: {
      title: string
      selectFile: string
      startImport: string
    }
    index: {
      title: string
      rebuildIndex: string
      indexStatus: string
    }
    duplicates: {
      title: string
      detectDuplicates: string
      noDuplicates: string
    }
    collections: {
      title: string
      createCollection: string
      noCollections: string
    }
    topics: {
      title: string
      buildTopics: string
      noTopics: string
    }
    review: {
      title: string
      reviewQueue: string
      noItems: string
    }
    aiTasks: {
      title: string
      noTasks: string
    }
    prompts: {
      title: string
      noPrompts: string
    }
    system: {
      title: string
      health: string
      stats: string
      doctor: string
      backup: string
      settings: string
    }
    account: {
      title: string
      profile: string
      preferences: string
    }
    security: {
      title: string
      sessions: string
      password: string
      activeSessions: string
      noSessions: string
      revokeAll: string
      revoke: string
    }
  }
  themes: {
    light: string
    dark: string
    system: string
  }
  locales: {
    enUS: string
    zhCN: string
    zhTW: string
  }
  backup: {
    title: string
    create: string
    creating: string
    download: string
    delete: string
    deleting: string
    noBackups: string
    confirmDelete: string
    size: string
    createdAt: string
  }
  settings: {
    title: string
    version: string
    storageProvider: string
    authEnabled: string
    searchEnabled: string
    aiEnabled: string
    databasePath: string
    storageRoot: string
    enabled: string
    disabled: string
    local: string
    qiniu: string
  }
}

export const enUS: Translations = {
  common: {
    save: 'Save',
    cancel: 'Cancel',
    delete: 'Delete',
    edit: 'Edit',
    create: 'Create',
    search: 'Search',
    loading: 'Loading...',
    error: 'Error',
    success: 'Success',
    confirm: 'Confirm',
  },
  nav: {
    vault: 'Vault',
    search: 'Search',
    import: 'Import',
    index: 'Index',
    duplicates: 'Duplicates',
    collections: 'Collections',
    topics: 'Topics',
    review: 'Review',
    aiTasks: 'AI Tasks',
    prompts: 'Prompts',
    system: 'System',
    settings: 'Settings',
    account: 'Account',
    security: 'Security',
    logout: 'Logout',
    login: 'Login',
  },
  auth: {
    signIn: 'Sign in to Cloud Vault',
    signOut: 'Sign out',
    username: 'Username',
    password: 'Password',
    currentPassword: 'Current Password',
    newPassword: 'New Password',
    confirmPassword: 'Confirm Password',
    changePassword: 'Change Password',
    loginFailed: 'Login failed',
    invalidCredentials: 'Invalid username or password',
    rateLimited: 'Too many login attempts, please try again later',
    updateProfile: 'Update Profile',
    displayName: 'Display Name',
    email: 'Email',
    locale: 'Language',
    theme: 'Theme',
    passwordMismatch: 'Passwords do not match',
    passwordTooShort: 'Password must be at least 10 characters',
    invalidCurrentPassword: 'Invalid current password',
    signingIn: 'Signing in...',
  },
  setup: {
    title: 'Initial Setup',
    description: 'Create your admin account to get started',
    createAccount: 'Create Admin Account',
    creating: 'Setting up...',
    setupFailed: 'Setup failed',
    alreadyInitialized: 'System already initialized',
  },
  pages: {
    vault: {
      title: 'Vault',
      documents: 'Documents',
      noDocuments: 'No documents yet',
      uploadDocument: 'Upload Document',
    },
    search: {
      title: 'Search',
      placeholder: 'Search documents...',
      noResults: 'No results found',
    },
    import: {
      title: 'Import',
      selectFile: 'Select File',
      startImport: 'Start Import',
    },
    index: {
      title: 'Index',
      rebuildIndex: 'Rebuild Index',
      indexStatus: 'Index Status',
    },
    duplicates: {
      title: 'Duplicates',
      detectDuplicates: 'Detect Duplicates',
      noDuplicates: 'No duplicates found',
    },
    collections: {
      title: 'Collections',
      createCollection: 'Create Collection',
      noCollections: 'No collections yet',
    },
    topics: {
      title: 'Topics',
      buildTopics: 'Build Topics',
      noTopics: 'No topics yet',
    },
    review: {
      title: 'Review',
      reviewQueue: 'Review Queue',
      noItems: 'No items to review',
    },
    aiTasks: {
      title: 'AI Tasks',
      noTasks: 'No AI tasks',
    },
    prompts: {
      title: 'Prompts',
      noPrompts: 'No prompts yet',
    },
    system: {
      title: 'System',
      health: 'Health',
      stats: 'Statistics',
      doctor: 'Doctor',
      backup: 'Backup',
      settings: 'Settings',
    },
    account: {
      title: 'Account',
      profile: 'Profile',
      preferences: 'Preferences',
    },
    security: {
      title: 'Security',
      sessions: 'Sessions',
      password: 'Password',
      activeSessions: 'active sessions',
      noSessions: 'No active sessions were returned.',
      revokeAll: 'Revoke All Other Sessions',
      revoke: 'Revoke',
    },
  },
  themes: {
    light: 'Light',
    dark: 'Dark',
    system: 'System',
  },
  locales: {
    enUS: 'English (US)',
    zhCN: '简体中文',
    zhTW: '繁體中文',
  },
  backup: {
    title: 'Backup Management',
    create: 'Create Backup',
    creating: 'Creating...',
    download: 'Download',
    delete: 'Delete',
    deleting: 'Deleting...',
    noBackups: 'No backups available',
    confirmDelete: 'Are you sure you want to delete this backup?',
    size: 'Size',
    createdAt: 'Created',
  },
  settings: {
    title: 'System Settings',
    version: 'Version',
    storageProvider: 'Storage Provider',
    authEnabled: 'Authentication',
    searchEnabled: 'Search',
    aiEnabled: 'AI Features',
    databasePath: 'Database Path',
    storageRoot: 'Storage Root',
    enabled: 'Enabled',
    disabled: 'Disabled',
    local: 'Local Storage',
    qiniu: 'Qiniu Cloud',
  },
}

export const zhCN: Translations = {
  common: {
    save: '保存',
    cancel: '取消',
    delete: '删除',
    edit: '编辑',
    create: '创建',
    search: '搜索',
    loading: '加载中...',
    error: '错误',
    success: '成功',
    confirm: '确认',
  },
  nav: {
    vault: '文档库',
    search: '搜索',
    import: '导入',
    index: '索引',
    duplicates: '重复项',
    collections: '集合',
    topics: '主题',
    review: '审阅',
    aiTasks: 'AI 任务',
    prompts: '提示词',
    system: '系统',
    settings: '设置',
    account: '账户',
    security: '安全',
    logout: '退出',
    login: '登录',
  },
  auth: {
    signIn: '登录到 Cloud Vault',
    signOut: '退出登录',
    username: '用户名',
    password: '密码',
    currentPassword: '当前密码',
    newPassword: '新密码',
    confirmPassword: '确认密码',
    changePassword: '修改密码',
    loginFailed: '登录失败',
    invalidCredentials: '用户名或密码错误',
    rateLimited: '登录尝试次数过多，请稍后再试',
    updateProfile: '更新资料',
    displayName: '显示名称',
    email: '邮箱',
    locale: '语言',
    theme: '主题',
    passwordMismatch: '两次输入的密码不匹配',
    passwordTooShort: '密码必须至少 10 个字符',
    invalidCurrentPassword: '当前密码错误',
    signingIn: '登录中...',
  },
  setup: {
    title: '初始设置',
    description: '创建管理员账户以开始使用',
    createAccount: '创建管理员账户',
    creating: '设置中...',
    setupFailed: '设置失败',
    alreadyInitialized: '系统已初始化',
  },
  pages: {
    vault: {
      title: '文档库',
      documents: '文档',
      noDocuments: '暂无文档',
      uploadDocument: '上传文档',
    },
    search: {
      title: '搜索',
      placeholder: '搜索文档...',
      noResults: '未找到结果',
    },
    import: {
      title: '导入',
      selectFile: '选择文件',
      startImport: '开始导入',
    },
    index: {
      title: '索引',
      rebuildIndex: '重建索引',
      indexStatus: '索引状态',
    },
    duplicates: {
      title: '重复项',
      detectDuplicates: '检测重复',
      noDuplicates: '未发现重复项',
    },
    collections: {
      title: '集合',
      createCollection: '创建集合',
      noCollections: '暂无集合',
    },
    topics: {
      title: '主题',
      buildTopics: '构建主题',
      noTopics: '暂无主题',
    },
    review: {
      title: '审阅',
      reviewQueue: '审阅队列',
      noItems: '无待审阅项',
    },
    aiTasks: {
      title: 'AI 任务',
      noTasks: '暂无 AI 任务',
    },
    prompts: {
      title: '提示词',
      noPrompts: '暂无提示词',
    },
    system: {
      title: '系统',
      health: '健康检查',
      stats: '统计',
      doctor: '诊断',
      backup: '备份',
      settings: '设置',
    },
    account: {
      title: '账户',
      profile: '个人资料',
      preferences: '偏好设置',
    },
    security: {
      title: '安全',
      sessions: '会话',
      password: '密码',
      activeSessions: '个活动会话',
      noSessions: '没有返回活动会话。',
      revokeAll: '撤销所有其他会话',
      revoke: '撤销',
    },
  },
  themes: {
    light: '浅色',
    dark: '深色',
    system: '跟随系统',
  },
  locales: {
    enUS: 'English (US)',
    zhCN: '简体中文',
    zhTW: '繁體中文',
  },
  backup: {
    title: '备份管理',
    create: '创建备份',
    creating: '创建中...',
    download: '下载',
    delete: '删除',
    deleting: '删除中...',
    noBackups: '暂无备份',
    confirmDelete: '确定要删除此备份吗？',
    size: '大小',
    createdAt: '创建时间',
  },
  settings: {
    title: '系统设置',
    version: '版本',
    storageProvider: '存储提供商',
    authEnabled: '认证',
    searchEnabled: '搜索',
    aiEnabled: 'AI 功能',
    databasePath: '数据库路径',
    storageRoot: '存储根目录',
    enabled: '已启用',
    disabled: '已禁用',
    local: '本地存储',
    qiniu: '七牛云',
  },
}

export const zhTW: Translations = {
  common: {
    save: '儲存',
    cancel: '取消',
    delete: '刪除',
    edit: '編輯',
    create: '建立',
    search: '搜尋',
    loading: '載入中...',
    error: '錯誤',
    success: '成功',
    confirm: '確認',
  },
  nav: {
    vault: '文件庫',
    search: '搜尋',
    import: '匯入',
    index: '索引',
    duplicates: '重複項',
    collections: '集合',
    topics: '主題',
    review: '審閱',
    aiTasks: 'AI 任務',
    prompts: '提示詞',
    system: '系統',
    settings: '設定',
    account: '帳戶',
    security: '安全',
    logout: '登出',
    login: '登入',
  },
  auth: {
    signIn: '登入到 Cloud Vault',
    signOut: '登出',
    username: '使用者名稱',
    password: '密碼',
    currentPassword: '目前密碼',
    newPassword: '新密碼',
    confirmPassword: '確認密碼',
    changePassword: '變更密碼',
    loginFailed: '登入失敗',
    invalidCredentials: '使用者名稱或密碼錯誤',
    rateLimited: '登入嘗試次數過多，請稍後再試',
    updateProfile: '更新資料',
    displayName: '顯示名稱',
    email: '電子郵件',
    locale: '語言',
    theme: '主題',
    passwordMismatch: '密碼不匹配',
    passwordTooShort: '密碼必須至少 10 個字元',
    invalidCurrentPassword: '當前密碼錯誤',
    signingIn: '登入中...',
  },
  setup: {
    title: '初始設定',
    description: '建立管理員帳戶以開始使用',
    createAccount: '建立管理員帳戶',
    creating: '設定中...',
    setupFailed: '設定失敗',
    alreadyInitialized: '系統已初始化',
  },
  pages: {
    vault: {
      title: '文件庫',
      documents: '文件',
      noDocuments: '暫無文件',
      uploadDocument: '上傳文件',
    },
    search: {
      title: '搜尋',
      placeholder: '搜尋文件...',
      noResults: '未找到結果',
    },
    import: {
      title: '匯入',
      selectFile: '選擇檔案',
      startImport: '開始匯入',
    },
    index: {
      title: '索引',
      rebuildIndex: '重建索引',
      indexStatus: '索引狀態',
    },
    duplicates: {
      title: '重複項',
      detectDuplicates: '偵測重複',
      noDuplicates: '未發現重複項',
    },
    collections: {
      title: '集合',
      createCollection: '建立集合',
      noCollections: '暫無集合',
    },
    topics: {
      title: '主題',
      buildTopics: '建構主題',
      noTopics: '暫無主題',
    },
    review: {
      title: '審閱',
      reviewQueue: '審閱佇列',
      noItems: '無待審閱項',
    },
    aiTasks: {
      title: 'AI 任務',
      noTasks: '暫無 AI 任務',
    },
    prompts: {
      title: '提示詞',
      noPrompts: '暫無提示詞',
    },
    system: {
      title: '系統',
      health: '健康檢查',
      stats: '統計',
      doctor: '診斷',
      backup: '備份',
      settings: '設定',
    },
    account: {
      title: '帳戶',
      profile: '個人資料',
      preferences: '偏好設定',
    },
    security: {
      title: '安全',
      sessions: '工作階段',
      password: '密碼',
      activeSessions: '個使用中的工作階段',
      noSessions: '沒有返回使用中的工作階段。',
      revokeAll: '撤銷所有其他工作階段',
      revoke: '撤銷',
    },
  },
  themes: {
    light: '淺色',
    dark: '深色',
    system: '跟隨系統',
  },
  locales: {
    enUS: 'English (US)',
    zhCN: '简体中文',
    zhTW: '繁體中文',
  },
  backup: {
    title: '備份管理',
    create: '建立備份',
    creating: '建立中...',
    download: '下載',
    delete: '刪除',
    deleting: '刪除中...',
    noBackups: '暫無備份',
    confirmDelete: '確定要刪除此備份嗎？',
    size: '大小',
    createdAt: '建立時間',
  },
  settings: {
    title: '系統設置',
    version: '版本',
    storageProvider: '儲存供應商',
    authEnabled: '認證',
    searchEnabled: '搜尋',
    aiEnabled: 'AI 功能',
    databasePath: '資料庫路徑',
    storageRoot: '儲存根目錄',
    enabled: '已啟用',
    disabled: '已停用',
    local: '本地儲存',
    qiniu: '七牛雲',
  },
}
