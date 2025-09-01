
export default {
  bootstrap: () => import('./main.server.mjs').then(m => m.default),
  inlineCriticalCss: true,
  baseHref: '/',
  locale: undefined,
  routes: [
  {
    "renderMode": 2,
    "route": "/"
  },
  {
    "renderMode": 2,
    "route": "/login"
  },
  {
    "renderMode": 2,
    "route": "/register"
  },
  {
    "renderMode": 2,
    "route": "/forgot-password"
  },
  {
    "renderMode": 2,
    "route": "/reset-password"
  },
  {
    "renderMode": 2,
    "route": "/oauth/callback"
  },
  {
    "renderMode": 2,
    "route": "/browse"
  }
],
  entryPointToBrowserMapping: undefined,
  assets: {
    'index.csr.html': {size: 7135, hash: '9ee332eecb6fca32dbd89d712bd6c44822a57886de07051261300bd8068d1561', text: () => import('./assets-chunks/index_csr_html.mjs').then(m => m.default)},
    'index.server.html': {size: 955, hash: '80c955fad5405557a6bfc737f477cc51e117f304e6f819733c2346d5d0f47fc9', text: () => import('./assets-chunks/index_server_html.mjs').then(m => m.default)},
    'oauth/callback/index.html': {size: 8768, hash: '2b8e650b2cb0f7474630cab49c72ec9bd44f7802af374c1d4e521a0024697b1c', text: () => import('./assets-chunks/oauth_callback_index_html.mjs').then(m => m.default)},
    'index.html': {size: 12976, hash: '4101b8ac552767ba95d42457edb5e6ed843b1449f367fcadff7a3cdeb097ddac', text: () => import('./assets-chunks/index_html.mjs').then(m => m.default)},
    'browse/index.html': {size: 16685, hash: 'db511cb61a9ee24e8c9864dfe19c385e091eae53a11b487eed77521a6e22f3bf', text: () => import('./assets-chunks/browse_index_html.mjs').then(m => m.default)},
    'login/index.html': {size: 16685, hash: 'db511cb61a9ee24e8c9864dfe19c385e091eae53a11b487eed77521a6e22f3bf', text: () => import('./assets-chunks/login_index_html.mjs').then(m => m.default)},
    'register/index.html': {size: 16461, hash: '06265bf805c1d65d667b35b5f4dbb5d4b48817f718f8a8f7d4c2cef07b3932e8', text: () => import('./assets-chunks/register_index_html.mjs').then(m => m.default)},
    'reset-password/index.html': {size: 15665, hash: 'ffc0480bc9459f77720819a34b37df77cc1f47b796511cc462d186637e2a7d91', text: () => import('./assets-chunks/reset-password_index_html.mjs').then(m => m.default)},
    'forgot-password/index.html': {size: 15158, hash: '2bfd3fe80cf724345f1eb983cc6b76010627416d6d3b27283d3562ec310abc12', text: () => import('./assets-chunks/forgot-password_index_html.mjs').then(m => m.default)},
    'styles-CBVIBG3V.css': {size: 635076, hash: 'tUStt+X1xxQ', text: () => import('./assets-chunks/styles-CBVIBG3V_css.mjs').then(m => m.default)}
  },
};
