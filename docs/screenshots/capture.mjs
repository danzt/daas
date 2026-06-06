/**
 * Captura screenshots de todas las páginas del manual DaaS usando Playwright.
 * Abre un browser VISIBLE — el usuario se loguea una vez y el script captura todo.
 *
 * Usage:
 *   cd /Users/danzt/Codes/daas
 *   node docs/screenshots/capture.mjs
 *
 * O con token directo (para correr headless):
 *   DAAS_TOKEN=xxx node docs/screenshots/capture.mjs
 */
import { chromium } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const BASE_URL = 'http://localhost:3000';
const OUT_DIR = __dirname;

const PAGES = [
  { name: 'dashboard',           path: '/dashboard',                 title: 'Dashboard' },
  { name: 'products',            path: '/products',                  title: 'Productos' },
  { name: 'products-categories', path: '/products/categories',       title: 'Categorías' },
  { name: 'inventory',           path: '/inventory',                 title: 'Inventario' },
  { name: 'sales-orders',        path: '/sales-orders',              title: 'Órdenes de venta' },
  { name: 'sales-orders-new',    path: '/sales-orders/new',          title: 'Nueva orden (POS)' },
  { name: 'shop-orders',         path: '/shop-orders',               title: 'Pedidos online' },
  { name: 'invoices',            path: '/invoices',                  title: 'Facturas internas' },
  { name: 'invoices-fiscal',     path: '/invoices/fiscal',           title: 'Facturas fiscales' },
  { name: 'suppliers',           path: '/suppliers',                 title: 'Proveedores' },
  { name: 'reports',             path: '/reports',                   title: 'Reportes' },
  { name: 'settings-payments',   path: '/settings/payment-methods', title: 'Métodos de pago' },
  { name: 'settings-branding',   path: '/settings/branding',        title: 'Branding' },
  { name: 'storefront',          path: '/t/daas-demo-store',         title: 'Storefront público' },
];

async function main() {
  const token = process.env.DAAS_TOKEN;
  const headless = !!token; // headless solo si tenemos token

  const browser = await chromium.launch({
    headless,
    args: headless ? [] : ['--window-size=1440,810'],
  });

  const context = await browser.newContext({
    viewport: { width: 1440, height: 810 },
  });
  const page = await context.newPage();

  if (token) {
    // Modo headless: inyectar token directamente
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await page.evaluate((tok) => localStorage.setItem('daas_token', tok), token);
    console.log('✓ Token inyectado (modo headless)');
  } else {
    // Modo visual: esperar que el usuario se loguee
    console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
    console.log('  Abriendo browser — loguéate en DaaS');
    console.log('  El script espera hasta que estés en /dashboard');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');

    await page.goto(`${BASE_URL}/auth/login`, { waitUntil: 'networkidle' });

    // Esperar hasta que localStorage tenga el token (login exitoso)
    await page.waitForFunction(
      () => !!localStorage.getItem('daas_token'),
      { timeout: 120_000 } // 2 minutos para que el usuario se loguee
    );
    console.log('✓ Login detectado — capturando páginas...\n');
  }

  for (const { name, path: pagePath, title } of PAGES) {
    process.stdout.write(`  Capturando: ${title.padEnd(30)}`);
    try {
      await page.goto(`${BASE_URL}${pagePath}`, { waitUntil: 'networkidle', timeout: 15_000 });
      await page.waitForTimeout(600); // dejar que charts y animaciones terminen
      const outPath = path.join(OUT_DIR, `${name}.jpg`);
      await page.screenshot({ path: outPath, type: 'jpeg', quality: 90, fullPage: false });
      console.log('✓');
    } catch (e) {
      console.log(`✗ (${e.message.slice(0, 60)})`);
    }
  }

  await browser.close();
  console.log('\n✅ Screenshots guardados en docs/screenshots/');
  console.log('   Próximo paso: git add docs/screenshots/*.jpg && git commit && git push');
}

main().catch(e => { console.error(e); process.exit(1); });
