# VaultHub 样式系统使用指南

## 设计哲学

简约大气 = 少即是多

- 用最少的颜色表达最丰富的语义
- 用统一的间距建立视觉秩序
- 用克制的动画提升体验
- 消除所有魔法数字，一切可追溯

## 快速开始

所有CSS变量已在 `variables.css` 中定义，并在 `main.js` 中全局引入。

### 使用示例

```vue
<template>
  <div class="card">
    <h3 class="card-title">标题</h3>
    <p class="card-text">内容</p>
  </div>
</template>

<style scoped>
.card {
  padding: var(--spacing-lg);
  background: var(--color-white);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  transition: all var(--duration-fast) var(--easing);
}

.card:hover {
  box-shadow: var(--shadow-md);
}

.card-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-md);
}

.card-text {
  font-size: var(--font-size-base);
  color: var(--color-text-secondary);
  line-height: var(--line-height-normal);
}
</style>
```

## 变量速查表

### 颜色

```css
/* 主色 */
var(--color-primary)          /* 主色 #667eea */
var(--color-primary-light)    /* 主色-浅 #7c8df0 */
var(--color-primary-dark)     /* 主色-深 #5568d3 */
var(--color-primary-bg)       /* 主色-背景 #f0f2ff */
var(--color-secondary)        /* 辅助色 #764ba2 */

/* 功能色 */
var(--color-success)          /* 成功 #10b981 */
var(--color-warning)          /* 警告 #f59e0b */
var(--color-error)            /* 错误 #ef4444 */
var(--color-info)             /* 信息 #3b82f6 */

/* 文字颜色 */
var(--color-text-primary)     /* 主要文字 #1f2937 */
var(--color-text-secondary)   /* 次要文字 #6b7280 */
var(--color-text-disabled)    /* 禁用文字 #9ca3af */

/* 边框和背景 */
var(--color-border)           /* 边框 #e5e7eb */
var(--color-bg)               /* 背景 #f9fafb */
var(--color-white)            /* 纯白 #ffffff */
```

### 间距（8px基础单位）

```css
var(--spacing-xs)    /* 4px */
var(--spacing-sm)    /* 8px */
var(--spacing-md)    /* 16px */
var(--spacing-lg)    /* 24px */
var(--spacing-xl)    /* 32px */
var(--spacing-2xl)   /* 48px */
var(--spacing-3xl)   /* 64px */
```

### 字体

```css
/* 字号 */
var(--font-size-xs)    /* 12px */
var(--font-size-sm)    /* 14px */
var(--font-size-base)  /* 16px */
var(--font-size-lg)    /* 18px */
var(--font-size-xl)    /* 24px */

/* 字重 */
var(--font-weight-normal)  /* 400 */
var(--font-weight-medium)  /* 500 */
var(--font-weight-bold)    /* 600 */

/* 行高 */
var(--line-height-tight)   /* 1.25 */
var(--line-height-normal)  /* 1.5 */
var(--line-height-loose)   /* 1.75 */
```

### 圆角

```css
var(--radius-sm)    /* 4px - 按钮、输入框 */
var(--radius-md)    /* 8px - 卡片 */
var(--radius-lg)    /* 12px - 弹窗 */
var(--radius-full)  /* 9999px - 全圆角 */
```

### 阴影

```css
var(--shadow-sm)  /* 轻微抬起 */
var(--shadow-md)  /* 卡片 */
var(--shadow-lg)  /* 弹窗 */
```

### 动画

```css
var(--duration-fast)  /* 150ms - 快速反馈 */
var(--duration-base)  /* 250ms - 标准过渡 */
var(--duration-slow)  /* 350ms - 慢速过渡 */
var(--easing)         /* cubic-bezier(0.4, 0, 0.2, 1) */
```

### 层级

```css
var(--z-dropdown)  /* 1000 */
var(--z-sticky)    /* 1020 */
var(--z-modal)     /* 1040 */
var(--z-popover)   /* 1060 */
var(--z-toast)     /* 1080 */
```

## 最佳实践

### 1. 禁止硬编码

```css
/* Bad */
.button {
  padding: 12px 20px;
  color: #667eea;
}

/* Good */
.button {
  padding: var(--spacing-sm) var(--spacing-lg);
  color: var(--color-primary);
}
```

### 2. 使用间距系统

```css
/* Bad */
.card {
  margin-bottom: 18px;
  padding: 22px;
}

/* Good */
.card {
  margin-bottom: var(--spacing-md);  /* 16px */
  padding: var(--spacing-lg);        /* 24px */
}
```

### 3. 统一动画

```css
/* Bad */
.button {
  transition: all 0.2s ease;
}

/* Good */
.button {
  transition: all var(--duration-fast) var(--easing);
}
```

### 4. 语义化颜色

```css
/* Bad */
.success-message {
  color: #10b981;
}

/* Good */
.success-message {
  color: var(--color-success);
}
```

## 响应式设计

使用媒体查询断点：

```css
/* 手机 */
@media (min-width: 640px) { }

/* 平板 */
@media (min-width: 768px) { }

/* 笔记本 */
@media (min-width: 1024px) { }

/* 桌面 */
@media (min-width: 1280px) { }
```

## 审查清单

提交前检查：

- [ ] 是否使用了魔法数字？
- [ ] 是否硬编码颜色？
- [ ] 间距是否是8的倍数？
- [ ] 是否过度使用动画？
- [ ] 响应式是否考虑？

记住：每个CSS规则都应该有明确的理由，否则删掉它。
