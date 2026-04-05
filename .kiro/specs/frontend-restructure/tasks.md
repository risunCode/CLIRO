# Implementation Plan

- [ ] 1. Kunci baseline source lama dan source aktif baru
  - Pastikan `frontend/src-old` diperlakukan sebagai referensi read-only dan tidak ikut menjadi dependency runtime source aktif.
  - Pastikan `frontend/src` menjadi satu-satunya root implementasi aktif untuk build, type check, dan alias path.
  - Tambahkan dokumentasi singkat struktur aktif vs legacy pada frontend.
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 9.1_

- [ ] 2. Bentuk scaffold `frontend/src` versi tengah
  - Buat struktur folder inti untuk `app`, `backend`, `features`, `shared`, `components/common`, `styles`, dan entrypoint utama.
  - Siapkan public surface minimum dan file placeholder yang diperlukan agar dependency graph baru dapat dibangun bertahap.
  - Sesuaikan alias path dan include config agar hanya menunjuk source aktif.
  - _Requirements: 2.1, 2.2, 2.3, 6.5, 9.1_

- [ ] 3. Ekstrak boundary backend dasar
  - Buat `frontend/src/backend/client` untuk wrapper Wails/runtime/fetch low-level.
  - Buat `frontend/src/backend/models` untuk model raw dan model backend-facing yang akan dipakai lintas layer.
  - Buat `frontend/src/backend/compat` untuk helper coercion, guard, dan mapper kompatibilitas.
  - Buat `frontend/src/backend/gateways` untuk kontrak domain-oriented yang dipakai app/features.
  - _Requirements: 2.2, 3.1, 3.2, 3.3, 3.4_

- [ ] 4. Migrasikan low-level Wails client dan runtime adapters
  - Pindahkan re-export binding generated Wails ke `backend/client` dengan naming yang jelas.
  - Pindahkan adapter runtime event/browser yang masih diperlukan ke boundary client yang tepat.
  - Pastikan tidak ada import langsung ke `wailsjs` dari `app/`, `features/`, atau `components`.
  - _Requirements: 3.1, 3.2, 6.3, 9.2_

- [ ] 5. Migrasikan helper kompatibilitas ke `backend/compat`
  - Pindahkan helper seperti coercion record/string/number/boolean serta mapper snake_case/camelCase dari pola lama ke `backend/compat`.
  - Pecah mapper per domain agar tidak tercampur antara accounts, router, logs, dan system.
  - Tambahkan surface ekspor yang jelas untuk dipakai gateway.
  - _Requirements: 3.2, 3.4, 8.2, 8.3_

- [ ] 6. Definisikan model backend lintas domain
  - Bentuk file model per domain di `backend/models` untuk system, logs, accounts, router, dan auth.
  - Normalisasi type yang sebelumnya tersebar di `app/types` dan `features/*/types` bila memang merupakan kontrak backend-facing.
  - Jaga agar type yang terekspos ke UI sudah konsisten dan tidak bergantung pada payload mentah.
  - _Requirements: 3.3, 7.2, 8.1_

- [ ] 7. Bangun gateway domain-oriented
  - Implementasikan `system-gateway`, `logs-gateway`, `accounts-gateway`, `auth-gateway`, dan `router-gateway` di atas client + compat.
  - Pastikan gateway mengembalikan shape data yang siap dipakai app/features dan mempertahankan behavior lama.
  - Tambahkan `index.ts` atau surface ekspor seragam untuk konsumsi layer atas.
  - _Requirements: 3.1, 3.3, 4.3, 7.2, 8.2_

- [ ] 8. Rapikan `app/` menjadi shell dan orchestration saja
  - Bangun ulang `App.svelte`, bootstrap, providers, routes/tabs, shell, overlays, dan services lintas feature di bawah struktur baru.
  - Pindahkan logic domain-spesifik keluar dari `app/` ke feature atau gateway yang tepat.
  - Pertahankan wiring UI dan lifecycle runtime yang sama dengan baseline.
  - _Requirements: 4.1, 4.2, 4.3, 7.1, 7.2_

- [ ] 9. Migrasikan shared primitives dan utilities lintas feature
  - Pindahkan komponen reusable netral domain ke `components/common` atau `shared/components` sesuai sifatnya.
  - Pindahkan store global non-domain dan helper umum ke `shared/stores` dan `shared/lib`.
  - Pastikan layer shared tidak bergantung pada feature tertentu.
  - _Requirements: 2.1, 6.1, 6.2, 8.1, 8.3_

- [ ] 10. Migrasikan feature `logs` sebagai vertikal pertama
  - Bangun ulang struktur `features/logs` dengan `components`, `lib`, `stores` atau `models` yang diperlukan serta `index.ts` public surface.
  - Hubungkan feature ke `logs-gateway` tanpa import ke binding/backend mentah.
  - Verifikasi workspace logs tetap menampilkan data dan aksi yang setara dengan baseline.
  - _Requirements: 5.1, 5.3, 6.2, 7.1, 7.3_

- [ ] 11. Migrasikan feature `router`
  - Pindahkan komponen proxy, scheduling, cloudflared, endpoint tester, model alias, dan CLI sync ke struktur feature baru.
  - Gantikan `features/router/api/*` lama dengan konsumsi `router-gateway` dan model yang sudah dinormalisasi.
  - Verifikasi seluruh kontrol API Router mempertahankan UI, copy, dan behavior sebelumnya.
  - _Requirements: 5.1, 5.3, 6.2, 7.1, 7.3, 9.3_

- [ ] 12. Migrasikan feature `accounts`
  - Pindahkan komponen workspace, modal, state, dan auth-session helpers ke boundary feature yang baru.
  - Pisahkan operasi account, quota, dan auth agar memakai `accounts-gateway` dan `auth-gateway`.
  - Verifikasi connect flow, refresh quota, toggle, delete, import, dan sync action tetap setara dengan baseline.
  - _Requirements: 5.1, 5.3, 6.2, 7.1, 7.3, 9.3_

- [ ] 13. Migrasikan feature `usage` dan feature residual lain
  - Pindahkan workspace usage dan helper formatting yang tersisa ke struktur feature baru.
  - Migrasikan `settings` atau modul residual lain hanya setelah pattern feature stabil.
  - Pastikan tidak ada logic domain yang tertinggal di `app/` atau `shared/` secara tidak semestinya.
  - _Requirements: 4.2, 5.1, 5.4, 7.1, 9.3_

- [ ] 14. Tegakkan naming convention dan public surface
  - Tambahkan `index.ts` pada boundary yang memang perlu diekspor dan hapus import deep-path yang melanggar boundary.
  - Samakan penamaan file/folder/komponen/service/gateway sesuai aturan desain.
  - Rapikan lokasi modul internal agar API publik tiap folder tetap jelas.
  - _Requirements: 5.4, 8.1, 8.2, 8.3, 8.4_

- [ ] 15. Tegakkan import rules pada source aktif
  - Audit dan perbaiki semua import agar mengikuti arah dependensi yang disepakati.
  - Hilangkan import dari `src-old`, import langsung ke `wailsjs`, dan cross-feature deep imports yang dilarang.
  - Jika perlu, tambahkan guard berbasis lint/checklist/script ringan untuk mendeteksi pelanggaran boundary utama.
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 16. Validasi parity dan finalisasi dokumentasi frontend
  - Jalankan `npm run check` dan validasi wiring utama setelah tiap fase besar serta pada kondisi akhir.
  - Lakukan smoke verification terhadap shell app, tabs/routes, accounts, router, logs, dan usage untuk memastikan tidak ada perubahan UI/behavior.
  - Perbarui `frontend/README.md` agar menjelaskan struktur akhir, boundary backend, naming, dan import rules.
  - _Requirements: 7.1, 7.2, 7.3, 8.1, 9.1, 9.3, 9.4_
