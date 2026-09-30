-- ============================================================
-- CBT SYSTEM - Script Darurat Reset Sesi Ujian
-- Jalankan via: docker exec -it cbt-db psql -U postgres -d cbt
-- ============================================================

-- ============================================================
-- BAGIAN 1: LIHAT STATUS SISWA
-- ============================================================

-- [1a] Lihat semua sesi aktif di jadwal tertentu (ganti nama jadwal)
-- SELECT es.id, es.status, es.submitted_at, u.full_name, sp.nis,
--        es.violation_count, es.total_score
-- FROM exam_sessions es
-- JOIN users u ON u.id = es.student_id
-- JOIN student_profiles sp ON sp.user_id = es.student_id
-- JOIN exam_schedules sch ON sch.id = es.schedule_id
-- WHERE sch.title ILIKE '%nama jadwal%'
-- ORDER BY u.full_name;

-- [1b] Lihat sesi siswa berdasarkan NIS
-- SELECT es.id, es.status, es.submitted_at, u.full_name, sp.nis,
--        es.violation_count, es.total_score
-- FROM exam_sessions es
-- JOIN users u ON u.id = es.student_id
-- JOIN student_profiles sp ON sp.user_id = es.student_id
-- WHERE sp.nis = 'GANTI_NIS_DISINI';

-- [1c] Lihat semua jadwal ujian yang ada
-- SELECT id, title, start_time, end_time, is_active FROM exam_schedules ORDER BY start_time DESC;


-- ============================================================
-- BAGIAN 2: RESET SESI SATU SISWA (berdasarkan NIS)
-- ============================================================
-- Ganti NIS_SISWA dengan NIS yang dituju

-- STEP 1: Hapus jawaban siswa
-- DELETE FROM student_answers
-- WHERE session_id IN (
--   SELECT es.id FROM exam_sessions es
--   JOIN student_profiles sp ON sp.user_id = es.student_id
--   WHERE sp.nis = 'NIS_SISWA'
-- );

-- STEP 2: Hapus sesi ujian
-- DELETE FROM exam_sessions
-- WHERE student_id = (
--   SELECT user_id FROM student_profiles WHERE nis = 'NIS_SISWA'
-- );


-- ============================================================
-- BAGIAN 3: RESET SESI SATU SISWA (berdasarkan session_id langsung)
-- ============================================================
-- Pakai ID yang didapat dari query BAGIAN 1

-- DELETE FROM student_answers WHERE session_id IN ('UUID1', 'UUID2');
-- DELETE FROM exam_sessions WHERE id IN ('UUID1', 'UUID2');


-- ============================================================
-- BAGIAN 4: RESET SEMUA SESI DI SATU JADWAL
-- ============================================================
-- HATI-HATI: Ini menghapus seluruh sesi & jawaban untuk 1 jadwal

-- DELETE FROM student_answers
-- WHERE session_id IN (
--   SELECT id FROM exam_sessions
--   WHERE schedule_id = (
--     SELECT id FROM exam_schedules WHERE title ILIKE '%nama jadwal%' LIMIT 1
--   )
-- );

-- DELETE FROM exam_sessions
-- WHERE schedule_id = (
--   SELECT id FROM exam_schedules WHERE title ILIKE '%nama jadwal%' LIMIT 1
-- );


-- ============================================================
-- BAGIAN 5: RESET TOKEN LOGIN HP SISWA (tanpa hapus jawaban)
-- ============================================================
-- Gunakan ini jika siswa ganti HP tapi jawaban tetap disimpan

-- UPDATE users SET session_token = ''
-- WHERE id = (
--   SELECT user_id FROM student_profiles WHERE nis = 'NIS_SISWA'
-- );


-- ============================================================
-- BAGIAN 6: LIHAT RINGKASAN STATUS UJIAN PER JADWAL
-- ============================================================
-- SELECT
--   sch.title AS jadwal,
--   COUNT(*) FILTER (WHERE es.status = 'SUBMITTED') AS selesai,
--   COUNT(*) FILTER (WHERE es.status = 'IN_PROGRESS') AS sedang_mengerjakan,
--   COUNT(*) FILTER (WHERE es.status = 'BLOCKED') AS terblokir,
--   COUNT(*) AS total_sesi
-- FROM exam_sessions es
-- JOIN exam_schedules sch ON sch.id = es.schedule_id
-- GROUP BY sch.title
-- ORDER BY sch.title;
