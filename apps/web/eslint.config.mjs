// @ts-check
import baseConfig from '@daas/config/eslint'

export default [
  ...baseConfig,
  {
    ignores: [
      '.nuxt/**',
      '.output/**',
      'node_modules/**',
      'dist/**',
    ],
  },
]
