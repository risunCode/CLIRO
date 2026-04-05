# Requirements Document

## Introduction

Restrukturisasi frontend CLIro-Go diperlukan untuk membangun ulang `frontend/src` dengan arsitektur yang lebih tegas tanpa mengubah UI, alur interaksi, data yang tampil, maupun perilaku runtime. Source lama dipertahankan di `frontend/src-old` sebagai referensi read-only, sementara source baru dibangun bertahap dengan boundary backend yang jelas, `app/` yang lebih rapi, serta migrasi feature satu per satu dengan aturan naming dan import yang ketat.

## Requirements

### Requirement 1

**User Story:** Sebagai maintainer frontend, saya ingin source lama dipindahkan dan diperlakukan sebagai referensi read-only, sehingga migrasi bisa dilakukan bertahap tanpa kehilangan baseline implementasi saat ini.

#### Acceptance Criteria

1. WHEN restrukturisasi dimulai THEN sistem SHALL mempertahankan implementasi lama di `frontend/src-old` sebagai sumber referensi read-only.
2. IF file berada di `frontend/src-old` THEN sistem SHALL memperlakukannya sebagai baseline historis yang tidak menjadi target pengembangan aktif.
3. WHEN source baru dibangun THEN sistem SHALL menggunakan `frontend/src` sebagai satu-satunya root implementasi aktif.
4. WHERE frontend source layout didokumentasikan sistem SHALL menjelaskan peran `frontend/src-old` dan `frontend/src` secara eksplisit.

### Requirement 2

**User Story:** Sebagai maintainer frontend, saya ingin `frontend/src` baru memiliki struktur versi tengah yang stabil, sehingga migrasi dapat dilakukan konsisten dan lebih mudah dipahami.

#### Acceptance Criteria

1. WHEN struktur baru dibuat THEN sistem SHALL menyediakan folder yang jelas untuk `app`, `features`, `shared`, `backend`, `components`, `styles`, dan entrypoint utama.
2. WHEN folder `backend` dibuat THEN sistem SHALL memisahkan tanggung jawab ke `backend/client`, `backend/gateways`, `backend/compat`, dan `backend/models`.
3. IF modul baru ditambahkan ke `frontend/src` THEN sistem SHALL ditempatkan pada boundary yang sesuai dengan tanggung jawabnya.
4. WHERE struktur baru digunakan sistem SHALL menghindari pengelompokan file berbasis kebiasaan lama yang mencampur UI, orchestration, dan backend access dalam satu area.

### Requirement 3

**User Story:** Sebagai developer frontend, saya ingin akses ke backend dipisahkan dengan boundary yang tegas, sehingga UI dan logic feature tidak lagi bergantung langsung pada binding Wails atau payload mentah.

#### Acceptance Criteria

1. WHEN UI atau feature membutuhkan data backend THEN sistem SHALL mengakses backend melalui gateway/domain adapter yang berada di bawah `frontend/src/backend`.
2. IF binding Wails atau payload mentah backend digunakan THEN sistem SHALL menempatkannya di `backend/client` atau `backend/compat`, bukan di layer feature/UI.
3. WHEN model frontend digunakan oleh app atau feature THEN sistem SHALL mengambil type/domain model dari `backend/models` atau model feature yang sudah dinormalisasi.
4. WHERE ada kebutuhan translasi camelCase, snake_case, atau bentuk payload historis sistem SHALL menanganinya di `backend/compat`.

### Requirement 4

**User Story:** Sebagai maintainer frontend, saya ingin `app/` hanya berisi shell dan orchestration tingkat aplikasi, sehingga boundary antara app shell dan feature menjadi jelas.

#### Acceptance Criteria

1. WHEN `app/` ditata ulang THEN sistem SHALL membatasi isinya pada bootstrap, shell, provider, route/tab composition, overlay, dan orchestration lintas feature.
2. IF suatu modul hanya relevan untuk satu domain bisnis THEN sistem SHALL menempatkannya di feature terkait, bukan di `app/`.
3. WHEN app-level state atau workflow lintas feature dibutuhkan THEN sistem SHALL menaruhnya di `app/` dengan kontrak yang jelas.
4. WHERE `app/` bergantung pada feature sistem SHALL menjaga dependensi satu arah dari `app/` ke public surface feature.

### Requirement 5

**User Story:** Sebagai developer frontend, saya ingin migrasi dilakukan per feature dengan naming yang jelas, sehingga perpindahan implementasi dapat diverifikasi tanpa mengubah behavior.

#### Acceptance Criteria

1. WHEN migrasi feature dilakukan THEN sistem SHALL memindahkan implementasi satu feature pada satu waktu dengan scope yang jelas.
2. IF feature belum dimigrasikan THEN sistem SHALL menjaga agar feature tersebut tetap dapat direferensikan dari baseline `frontend/src-old` hanya untuk keperluan baca/manual porting, bukan import runtime.
3. WHEN feature selesai dimigrasikan THEN sistem SHALL menyediakan public surface yang jelas untuk digunakan `app/` atau feature lain.
4. WHERE naming file, folder, dan modul ditentukan sistem SHALL menggunakan nama yang deskriptif dan konsisten terhadap domain.

### Requirement 6

**User Story:** Sebagai maintainer codebase, saya ingin aturan import yang tegas, sehingga dependency graph tetap sehat selama dan setelah migrasi.

#### Acceptance Criteria

1. WHEN modul di `frontend/src` mengimpor dependensi THEN sistem SHALL mengikuti aturan import berbasis boundary yang terdokumentasi.
2. IF modul berada di `features/*` THEN sistem SHALL melarang import langsung ke implementasi internal feature lain kecuali melalui public surface yang disetujui.
3. IF modul berada di `features/*`, `app/*`, atau `components/*` THEN sistem SHALL melarang import langsung ke `wailsjs`.
4. IF modul aktif di `frontend/src` mencoba mengimpor dari `frontend/src-old` THEN sistem SHALL menganggapnya pelanggaran arsitektur.
5. WHERE alias path digunakan sistem SHALL menjaga alias tersebut menunjuk ke source aktif dan konsisten dengan boundary baru.

### Requirement 7

**User Story:** Sebagai pengguna aplikasi, saya ingin restrukturisasi tidak mengubah UI maupun behavior, sehingga upgrade internal tidak mempengaruhi pengalaman penggunaan.

#### Acceptance Criteria

1. WHEN restrukturisasi selesai THEN sistem SHALL mempertahankan tampilan, copy, urutan interaksi, dan hasil operasi yang setara dengan baseline sebelum restrukturisasi.
2. IF ada perubahan internal pada state, adapter, atau struktur folder THEN sistem SHALL tetap menjaga kontrak UI dan perilaku runtime yang ada.
3. WHEN feature dimigrasikan THEN sistem SHALL memverifikasi parity terhadap baseline sebelum melanjutkan ke feature berikutnya.
4. WHERE perbedaan implementation detail diperlukan sistem SHALL membatasi perubahan hanya pada struktur internal dan maintainability.

### Requirement 8

**User Story:** Sebagai maintainer frontend, saya ingin naming convention dan public surface tiap layer terdokumentasi, sehingga kontribusi lanjutan mengikuti pola yang sama.

#### Acceptance Criteria

1. WHEN struktur baru didokumentasikan THEN sistem SHALL mendefinisikan naming convention untuk folder, file, komponen Svelte, store, service, mapper, dan gateway.
2. WHEN sebuah folder memiliki boundary publik THEN sistem SHALL menentukan file masuk/public surface yang eksplisit.
3. IF modul bersifat internal THEN sistem SHALL menempatkannya di lokasi yang tidak mengaburkan API publiknya.
4. WHERE aturan ini berlaku sistem SHALL mendeskripsikan contoh penggunaan yang benar dan yang dilarang.

### Requirement 9

**User Story:** Sebagai maintainer proyek, saya ingin rencana migrasi dapat dieksekusi bertahap dan aman, sehingga codebase tetap buildable sepanjang proses.

#### Acceptance Criteria

1. WHEN implementasi dilakukan THEN sistem SHALL membagi migrasi ke fase-fase inkremental yang menjaga frontend tetap dapat di-build dan di-check.
2. IF sebuah fase memperkenalkan scaffold baru THEN sistem SHALL menyediakan compatibility layer seperlunya untuk menjaga migrasi bertahap.
3. WHEN fase migrasi selesai THEN sistem SHALL memiliki kriteria selesai yang dapat diverifikasi.
4. WHERE validasi dilakukan sistem SHALL menggunakan pemeriksaan otomatis yang relevan seperti type check dan smoke validation terhadap wiring utama.
