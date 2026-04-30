<script>
  let { onGetStarted } = $props()
</script>

<div class="landing">

  <nav>
    <div class="nav-in">
      <span class="logo">Bajet</span>
      <button class="ncta" onclick={onGetStarted}>Mulai Sekarang</button>
    </div>
  </nav>

  <section class="hero wrap">
    <div class="badge"><span class="dot"></span> Gratis · Tanpa akun bank</div>
    <h1>Budget tanpa<br/><em>catat apa-apa.</em></h1>
    <p class="hsub">Tetapkan budget per periode. Check-in sekali saat selesai.<br/><strong>Tanpa transaksi. Tanpa kategori. Tanpa rasa bersalah.</strong></p>

    <button class="btn-hero" onclick={onGetStarted}>Mulai Gratis →</button>

    <div class="prev">
      <div class="ph">
        <div>
          <div class="pm">Mei · Cycle 3</div>
          <div class="pb">Rp 2.000.000 / cycle</div>
        </div>
        <div style="font-size:.70rem;color:var(--text-muted)">4 periode</div>
      </div>
      <div class="pc"><div><div class="pl">Periode 1</div><div class="pd">25 Apr – 2 Mei</div></div><div><div class="pk ks">Sisa</div><div class="pr rg">+Rp 65.000</div></div></div>
      <div class="pc"><div><div class="pl">Periode 2</div><div class="pd">3 – 10 Mei</div></div><div><div class="pk ks">Sisa</div><div class="pr rg">+Rp 12.500</div></div></div>
      <div class="pc"><div><div class="pl">Periode 3</div><div class="pd">11 – 17 Mei</div></div><div><div class="pk ko">Defisit</div><div class="pr rr">−Rp 30.000</div></div></div>
      <div class="pc open-pc"><div><div class="pl">Periode 4 · Sekarang</div><div class="pd">18 – 24 Mei</div></div><div><div class="pk kop">Aktif</div></div></div>
      <button class="ckin" onclick={onGetStarted}>✓ &nbsp;Check-in sekarang &nbsp;(30 detik)</button>
    </div>
  </section>

  <section class="alt sec">
    <div class="wrap">
      <span class="sl">Terasa familiar?</span>
      <h2>Setiap app budgeting<br/>gagal dengan cara yang sama.</h2>
      <div class="pg">
        <div class="pca"><div class="pi">😤</div><h3>Jebakan YNAB</h3><p>Setup-nya sempurna. Sembilan hari kemudian kamu sudah dua minggu ketinggalan dan malas membukanya lagi.</p></div>
        <div class="pca"><div class="pi">📊</div><h3>Kuburan Spreadsheet</h3><p>Ada tab bertuliskan "Oktober." Sekarang sudah Maret. Kamu terus berniat kembali tapi makin lama makin berat.</p></div>
        <div class="pca"><div class="pi">🎨</div><h3>Kebohongan Template Notion</h3><p>Tampilannya cantik. Dibuka dua kali. Lalu kehidupan berjalan dan mencatat 30 transaksi seminggu tidak terasa menyenangkan lagi.</p></div>
      </div>
      <p style="margin-top:24px;font-size:.91rem;max-width:500px">Masalahnya bukan disiplinmu. <strong style="color:var(--text)">Tapi desainnya.</strong> App budgeting dibuat untuk akuntan, bukan manusia biasa.</p>
    </div>
  </section>

  <section class="sec wrap">
    <span class="sl">Cara kerjanya</span>
    <h2>Tiga langkah. Satu check-in.<br/>Itu saja produknya.</h2>
    <div class="sg">
      <div class="st"><div class="sn">1</div><h3>Tetapkan budget cycle</h3><p>Satu total untuk satu cycle. App otomatis membaginya jadi 4 periode — rata atau disesuaikan perilakumu.</p><span class="stm">⏱ 1 menit, sekali saja</span></div>
      <div class="st"><div class="sn">2</div><h3>Jalani hidupmu</h3><p>Tidak perlu catat. Tidak perlu log. Tidak perlu kategori. Jangan buka app sampai periode berakhir.</p><span class="stm">⏱ 0 menit</span></div>
      <div class="st"><div class="sn">3</div><h3>Check-in sekali</h3><p>Masukkan satu angka: berapa yang tersisa (sisa) atau berapa yang kurang (defisit). Selesai.</p><span class="stm">⏱ 30 detik</span></div>
    </div>
  </section>

  <section class="alt sec">
    <div class="wrap">
      <span class="sl">Perbedaannya</span>
      <h2>Cara lama vs. cara ini.</h2>
      <div class="cg">
        <div class="cc"><h3>😩 App lainnya</h3>
          <div class="ci"><span>❌</span><span>Hubungkan rekening bank &amp; berikan akses baca</span></div>
          <div class="ci"><span>❌</span><span>Kategorikan 200+ transaksi setiap bulan</span></div>
          <div class="ci"><span>❌</span><span>Frustrasi saat koneksi bank terputus lagi</span></div>
          <div class="ci"><span>❌</span><span>Merasa bersalah kalau absen dua hari</span></div>
          <div class="ci"><span>❌</span><span>Berhenti di minggu kedua dan menyalahkan diri sendiri</span></div>
        </div>
        <div class="cc ours"><h3>✦ Bajet</h3>
          <div class="ci"><span>✅</span><span>Tidak perlu koneksi bank. Kami tidak pernah melihat transaksimu.</span></div>
          <div class="ci"><span>✅</span><span>Satu angka per periode. Itu satu-satunya input-mu.</span></div>
          <div class="ci"><span>✅</span><span>Tidak ada yang tertinggal — check-in kapan kamu siap.</span></div>
          <div class="ci"><span>✅</span><span>Lewat satu periode? Lanjut yang berikutnya. Tanpa rasa bersalah.</span></div>
          <div class="ci"><span>✅</span><span>Masih pakai budgeting 6 bulan kemudian karena memang berhasil.</span></div>
        </div>
      </div>
    </div>
  </section>

  <section class="sec wrap">
    <span class="sl">Pengguna awal</span>
    <h2>Untuk orang yang<br/>sudah menyerah dengan app budgeting.</h2>
    <div class="qg">
      <div class="qc"><div class="qs">★★★★★</div><p class="qt">"Saya hapus YNAB enam bulan lalu dan berjanji tidak akan membuka app budget lagi. Ini yang pertama dirancang untuk cara hidup saya — bukan cara hidup yang seharusnya."</p><div class="qa"><div class="qav">A</div><div><div class="qn">Alex R.</div><div class="ql">Berlin, Germany · Beta tester</div></div></div></div>
      <div class="qc"><div class="qs">★★★★★</div><p class="qt">"Saya freelancer dengan penghasilan tidak tetap. Sistem berbasis periode memang cara manusia berpikir. Kamu habiskan sampai gajian berikutnya, bukan sampai hari Selasa."</p><div class="qa"><div class="qav">M</div><div><div class="qn">Maya T.</div><div class="ql">Toronto, Canada · Beta tester</div></div></div></div>
      <div class="qc"><div class="qs">★★★★★</div><p class="qt">"Konsep sisa/defisit sangat jujur. Tidak pura-pura sempurna — hanya lacak apakah kamu bertahan di periode ini. Itu sudah cukup."</p><div class="qa"><div class="qav">D</div><div><div class="qn">Dito S.</div><div class="ql">Jakarta, Indonesia · Beta tester</div></div></div></div>
    </div>
  </section>

  <section class="ps">
    <div class="psi">
      <div class="pit">🔒 Tidak perlu koneksi bank.</div>
      <div class="pit">🧾 Kami tidak pernah melihat transaksimu.</div>
      <div class="pit">🚫 Tanpa iklan. Tanpa jual data.</div>
      <div class="pit">✨ Gratis untuk 1 budget.</div>
    </div>
  </section>

  <section class="fc">
    <h2>Siap berhenti mencatat?</h2>
    <p>Mulai gratis — tidak perlu kartu kredit.</p>
    <button class="btn-hero" onclick={onGetStarted}>Mulai Sekarang →</button>
  </section>

  <footer>
    <p style="margin-bottom:5px"><strong>Bajet</strong> &nbsp;·&nbsp; Budget per periode, bukan per struk.</p>
    <p>© 2026 Bajet &nbsp;·&nbsp; Dibuat oleh manusia yang juga berhenti pakai YNAB</p>
  </footer>

</div>

<style>
  /* ── Reset & base ── */
  .landing {
    --bg: #0c0c0e;
    --bg2: #111115;
    --bg3: #18181e;
    --border: rgba(255,255,255,.07);
    --lt: #f0f0f4;
    --muted: #7c7c8a;
    --accent: #154374;
    --accent2: #2075D0;
    --glow: rgba(33,117,208,.25);
    --green: #3ecf8e;
    --red: #e05c5c;
    background: var(--bg);
    color: var(--lt);
    font-family: 'Albert Sans', system-ui, sans-serif;
    -webkit-font-smoothing: antialiased;
    overflow-x: hidden;
  }

  h1 { font-size: clamp(2rem,6vw,3.5rem); font-weight: 800; line-height: 1.1; letter-spacing: -.03em; }
  h2 { font-size: clamp(1.4rem,4vw,2.1rem); font-weight: 700; line-height: 1.2; letter-spacing: -.02em; }
  h3 { font-size: 1rem; font-weight: 600; }
  p  { color: var(--muted); }

  .wrap { max-width: 860px; margin: 0 auto; padding: 0 20px; }
  .sec  { padding: 80px 0; }

  /* ── Nav ── */
  nav {
    position: sticky;
    top: 0;
    z-index: 100;
    background: rgba(12,12,14,.88);
    backdrop-filter: blur(18px);
    border-bottom: 1px solid var(--border);
  }
  .nav-in {
    max-width: 860px;
    margin: 0 auto;
    padding: 0 20px;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .logo {
    font-size: 1.1rem;
    font-weight: 800;
    letter-spacing: -.04em;
    background: linear-gradient(135deg,#fff 30%,#F2E942);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }
  .ncta {
    background: var(--accent);
    color: #fff;
    border: none;
    border-radius: 8px;
    padding: 8px 18px;
    font-size: .85rem;
    font-weight: 600;
    cursor: pointer;
    transition: .2s;
    font-family: inherit;
  }
  .ncta:hover { background: var(--accent2); transform: translateY(-1px); }

  /* ── Hero ── */
  .hero {
    padding: 88px 0 64px;
    text-align: center;
    position: relative;
    overflow: hidden;
  }
  .hero::before {
    content: '';
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse 70% 50% at 50% 0%, var(--glow), transparent 70%);
    pointer-events: none;
  }
  .badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: rgba(33,117,208,.12);
    border: 1px solid rgba(33,117,208,.3);
    color: #7aabee;
    font-size: .72rem;
    font-weight: 700;
    letter-spacing: .05em;
    padding: 5px 14px;
    border-radius: 100px;
    margin-bottom: 24px;
    text-transform: uppercase;
  }
  .dot { width: 6px; height: 6px; border-radius: 50%; background: var(--accent2); display: inline-block; }
  .hero h1 { margin-bottom: 16px; }
  .hero h1 em { font-style: normal; color: #F2E942; }
  .hsub {
    font-size: clamp(1rem,2.5vw,1.15rem);
    color: var(--muted);
    max-width: 520px;
    margin: 0 auto 34px;
    line-height: 1.7;
  }
  .hsub strong { color: var(--lt); font-weight: 500; }

  .btn-hero {
    background: var(--accent2);
    color: #fff;
    border: none;
    border-radius: 10px;
    padding: 14px 28px;
    font-size: 1rem;
    font-weight: 700;
    cursor: pointer;
    transition: .2s;
    font-family: inherit;
    margin-bottom: 40px;
  }
  .btn-hero:hover { background: var(--accent); transform: translateY(-2px); box-shadow: 0 8px 30px var(--glow); }

  /* ── Preview card ── */
  .prev {
    margin: 0 auto;
    max-width: 310px;
    background: var(--bg3);
    border: 1px solid var(--border);
    border-radius: 22px;
    padding: 20px;
    box-shadow: 0 40px 80px rgba(0,0,0,.5);
    text-align: left;
  }
  .ph { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
  .pm { font-size: .72rem; color: var(--muted); }
  .pb { font-size: .96rem; font-weight: 700; color: var(--lt); }
  .pc {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: 11px;
    padding: 12px 14px;
    margin-bottom: 8px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .open-pc { border-color: rgba(33,117,208,.3); }
  .pl { font-size: .68rem; color: var(--muted); margin-bottom: 1px; }
  .pd { font-size: .86rem; font-weight: 600; color: var(--lt); }
  .pk { font-size: .68rem; font-weight: 700; padding: 3px 8px; border-radius: 100px; }
  .ks  { background: rgba(62,207,142,.12); color: var(--green); }
  .ko  { background: rgba(224,92,92,.12);  color: var(--red); }
  .kop { background: rgba(33,117,208,.15); color: #7aabee; }
  .pr  { font-size: .80rem; font-weight: 600; margin-top: 2px; text-align: right; }
  .rg  { color: var(--green); }
  .rr  { color: var(--red); }
  .ckin {
    width: 100%;
    margin-top: 12px;
    background: linear-gradient(135deg, var(--accent2), var(--accent));
    color: #fff;
    border: none;
    border-radius: 10px;
    padding: 10px;
    font-size: .86rem;
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-family: inherit;
  }

  /* ── Alt section ── */
  .alt {
    background: var(--bg2);
    border-top: 1px solid var(--border);
    border-bottom: 1px solid var(--border);
  }
  .sl {
    font-size: .70rem;
    font-weight: 700;
    letter-spacing: .1em;
    text-transform: uppercase;
    color: var(--accent2);
    margin-bottom: 8px;
    display: block;
  }

  /* Pain grid */
  .pg { display: grid; grid-template-columns: repeat(auto-fit,minmax(220px,1fr)); gap: 13px; margin-top: 32px; }
  .pca { background: var(--bg3); border: 1px solid var(--border); border-radius: 14px; padding: 18px 20px; }
  .pi { font-size: 1.3rem; margin-bottom: 9px; }
  .pca h3 { margin-bottom: 5px; font-size: .9rem; color: var(--lt); }
  .pca p  { font-size: .84rem; }

  /* Steps grid */
  .sg { display: grid; grid-template-columns: repeat(auto-fit,minmax(200px,1fr)); gap: 18px; margin-top: 36px; }
  .st { background: var(--bg3); border: 1px solid var(--border); border-radius: 14px; padding: 24px 20px; }
  .sn {
    width: 32px; height: 32px; border-radius: 9px;
    background: linear-gradient(135deg, var(--accent2), var(--accent));
    display: flex; align-items: center; justify-content: center;
    font-size: .80rem; font-weight: 800; color: #fff; margin-bottom: 12px;
  }
  .st h3 { margin-bottom: 6px; color: var(--lt); }
  .st p  { font-size: .84rem; }
  .stm {
    margin-top: 11px;
    display: inline-flex; align-items: center; gap: 4px;
    font-size: .70rem; font-weight: 600; color: var(--green);
    background: rgba(62,207,142,.1); padding: 3px 8px; border-radius: 100px;
  }

  /* Comparison */
  .cg { display: grid; grid-template-columns: 1fr 1fr; gap: 13px; margin-top: 32px; }
  .cc { background: var(--bg3); border: 1px solid var(--border); border-radius: 14px; padding: 22px 18px; }
  .cc.ours { border-color: rgba(33,117,208,.35); background: rgba(33,117,208,.06); }
  .cc h3 { font-size: .86rem; margin-bottom: 16px; color: var(--muted); }
  .cc.ours h3 { color: var(--accent2); }
  .ci { display: flex; align-items: flex-start; gap: 8px; margin-bottom: 11px; font-size: .83rem; }
  .ci:last-child { margin-bottom: 0; }
  .ci span { color: var(--muted); line-height: 1.5; }
  .cc.ours .ci span { color: var(--lt); }

  /* Quotes */
  .qg { display: grid; grid-template-columns: repeat(auto-fit,minmax(240px,1fr)); gap: 13px; margin-top: 32px; }
  .qc { background: var(--bg3); border: 1px solid var(--border); border-radius: 14px; padding: 20px; }
  .qs { color: #f4b740; font-size: .80rem; margin-bottom: 9px; }
  .qt { font-size: .86rem; color: var(--lt); line-height: 1.65; margin-bottom: 11px; }
  .qa { display: flex; align-items: center; gap: 8px; }
  .qav {
    width: 28px; height: 28px; border-radius: 50%;
    background: linear-gradient(135deg, var(--accent2), var(--green));
    display: flex; align-items: center; justify-content: center;
    font-size: .70rem; font-weight: 800; color: #fff; flex-shrink: 0;
  }
  .qn { font-size: .74rem; color: var(--lt); font-weight: 600; }
  .ql { font-size: .70rem; color: var(--muted); }

  /* Privacy strip */
  .ps { background: var(--bg2); border-top: 1px solid var(--border); border-bottom: 1px solid var(--border); }
  .psi { max-width: 860px; margin: 0 auto; padding: 26px 20px; display: flex; gap: 24px; flex-wrap: wrap; justify-content: center; }
  .pit { display: flex; align-items: center; gap: 8px; font-size: .82rem; font-weight: 500; color: var(--lt); }

  /* Final CTA */
  .fc {
    text-align: center;
    padding: 90px 20px;
    position: relative;
    overflow: hidden;
  }
  .fc::before {
    content: '';
    position: absolute; inset: 0;
    background: radial-gradient(ellipse 60% 60% at 50% 100%, var(--glow), transparent 70%);
    pointer-events: none;
  }
  .fc h2 { margin-bottom: 12px; color: var(--lt); }
  .fc > p { margin-bottom: 28px; max-width: 400px; margin-left: auto; margin-right: auto; }
  .fc .btn-hero { margin-bottom: 0; }

  footer {
    border-top: 1px solid var(--border);
    padding: 24px 20px;
    text-align: center;
    font-size: .76rem;
    color: var(--muted);
  }
  footer strong { color: var(--lt); }

  /* Animations */
  @keyframes fu { from { opacity: 0; transform: translateY(16px); } to { opacity: 1; transform: translateY(0); } }
  .hero > * { animation: fu .55s ease both; }
  .hero > *:nth-child(1) { animation-delay: .05s; }
  .hero > *:nth-child(2) { animation-delay: .13s; }
  .hero > *:nth-child(3) { animation-delay: .20s; }
  .hero > *:nth-child(4) { animation-delay: .27s; }
  .hero > *:nth-child(5) { animation-delay: .34s; }

  @media (max-width: 560px) {
    .hero  { padding: 56px 0 40px; }
    .sec   { padding: 56px 0; }
    .cg    { grid-template-columns: 1fr; }
    .btn-hero { width: 100%; }
  }
</style>
