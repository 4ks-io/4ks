import nextConfig from 'eslint-config-next/core-web-vitals';

// react-hooks/set-state-in-effect, static-components, immutability, purity are
// new in eslint-plugin-react-hooks@5 (shipped with Next 16) and enforce React
// Compiler compatibility. Disabled here for the Next 16 baseline because
// recipe-context.tsx and other files have legitimate mutex/side-effect patterns
// that need targeted review before React Compiler is enabled. Track in:
// apps/web/src/providers/recipe-context/recipe-context.tsx
const reactCompilerRuleOverrides = {
  'react-hooks/set-state-in-effect': 'off',
  'react-hooks/static-components': 'off',
  'react-hooks/immutability': 'off',
  'react-hooks/purity': 'off',
};

export default [
  ...nextConfig,
  {
    rules: reactCompilerRuleOverrides,
  },
];
