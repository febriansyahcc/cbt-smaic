// Membuat PDF buku panduan dari file HTML di docs/manual.
// Pemakaian: cd docs/manual/build && npm install && node build.mjs [nama-file.html ...]
import { chromium } from 'playwright';
import { PDFDocument } from 'pdf-lib';
import { readdirSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const files = process.argv.slice(2).length ? process.argv.slice(2) : readdirSync(root).filter((f) => /^panduan-.*\.html$/.test(f));

const header = (title) => `
  <div style="width:100%;padding:0 18mm;font-family:Inter,system-ui,sans-serif;font-size:7.5pt;color:#64748b;display:flex;justify-content:space-between;">
    <span><b style="color:#0f172a">${title}</b> · CBT SMAS Islamic Centre Demak</span>
  </div>`;
const footer = `
  <div style="width:100%;padding:0 18mm;font-family:Inter,system-ui,sans-serif;font-size:7.5pt;color:#64748b;display:flex;justify-content:space-between;">
    <span>SMAS Islamic Centre Demak</span><span>Halaman <span class="pageNumber"></span></span>
  </div>`;

const browser = await chromium.launch();
for (const file of files) {
  const page = await browser.newPage();
  await page.goto('file://' + join(root, file), { waitUntil: 'networkidle' }).catch(() => {});
  await page.evaluate(() => document.fonts.ready);
  const title = await page.title();

  await page.evaluate(() => document.body.classList.add('print-cover'));
  const cover = await page.pdf({ format: 'A4', printBackground: true, pageRanges: '1', margin: { top: 0, bottom: 0, left: 0, right: 0 } });

  await page.evaluate(() => { document.body.classList.remove('print-cover'); document.body.classList.add('print-body'); });
  const body = await page.pdf({
    format: 'A4', printBackground: true, displayHeaderFooter: true,
    headerTemplate: header(title), footerTemplate: footer,
    margin: { top: '20mm', bottom: '18mm', left: '18mm', right: '18mm' },
  });

  const out = await PDFDocument.create();
  for (const bytes of [cover, body]) {
    const src = await PDFDocument.load(bytes);
    for (const p of await out.copyPages(src, src.getPageIndices())) out.addPage(p);
  }
  out.setTitle(title);
  out.setAuthor('SMAS Islamic Centre Demak');
  const target = join(root, 'pdf', file.replace(/\.html$/, '.pdf'));
  writeFileSync(target, await out.save());
  console.log(`${file} -> pdf/${file.replace(/\.html$/, '.pdf')} (${out.getPageCount()} halaman)`);
  await page.close();
}
await browser.close();
