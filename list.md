# BillionMail – Danh sách issues cần/nên cập nhật

Nguồn: https://github.com/Billionmail/BillionMail/issues (đọc toàn bộ **232 issue đang mở**, ngày 2026-10-23).

Tiêu chí chọn: (1) lỗi bảo mật, (2) lỗi làm hỏng chức năng cốt lõi (gửi mail, thống kê, webmail),
(3) lỗi có nguyên nhân rõ ràng trong code, (4) tính năng nhỏ – giá trị cao. Các issue thuần hỏi đáp
(DNS, cài đặt, cổng 25 bị chặn…) hoặc yêu cầu kiến trúc rất lớn (HA đa máy, drag-and-drop automation…)
không đưa vào đợt này.

Trạng thái: ✅ đã code + có test · ⚠️ đã xử lý một phần (ghi rõ lý do)

## Phần 1 – 20 issues ưu tiên

| # | Issue | Loại | Mức độ | Trạng thái |
|---|-------|------|--------|-----------|
| 1 | [#390](https://github.com/Billionmail/BillionMail/issues/390) Token tracker `/pmta/*` giả mạo được + open redirect | Bảo mật | Nghiêm trọng | ✅ |
| 2 | [#364](https://github.com/Billionmail/BillionMail/issues/364) / [#381](https://github.com/Billionmail/BillionMail/issues/381) / [#387](https://github.com/Billionmail/BillionMail/issues/387) / [#338](https://github.com/Billionmail/BillionMail/issues/338) / [#361](https://github.com/Billionmail/BillionMail/issues/361) / [#360](https://github.com/Billionmail/BillionMail/issues/360) SafePath chặn nhầm Roundcube, ACME, link tracking | Bug | Cao | ✅ |
| 3 | [#172](https://github.com/Billionmail/BillionMail/issues/172) Đọc file tùy ý qua `/file/read`, `/file/download` | Bảo mật | Cao | ✅ |
| 4 | [#355](https://github.com/Billionmail/BillionMail/issues/355) Nhiều SMTP relay bị trộn credential giữa các domain | Bug | Cao | ✅ |
| 5 | [#379](https://github.com/Billionmail/BillionMail/issues/379) Test kết nối SMTP relay treo vô hạn ở cổng TLS không chuẩn (2465) | Bug | Cao | ✅ |
| 6 | [#376](https://github.com/Billionmail/BillionMail/issues/376) Display Name của task bị bỏ qua trong header From | Bug | Cao | ✅ |
| 7 | [#389](https://github.com/Billionmail/BillionMail/issues/389) SPF sinh sai IP khi domain dùng IP riêng (multi-IP) | Bug | Trung bình | ✅ |
| 8 | [#378](https://github.com/Billionmail/BillionMail/issues/378) Import tạo mailbox username rỗng & không xoá được | Bug | Trung bình | ✅ |
| 9 | [#350](https://github.com/Billionmail/BillionMail/issues/350) Mật khẩu tạo hàng loạt không đúng chính sách | Bug | Trung bình | ✅ |
| 10 | [#348](https://github.com/Billionmail/BillionMail/issues/348) Không bật/tắt được trạng thái mailbox từ danh sách | Bug | Trung bình | ✅ |
| 11 | [#344](https://github.com/Billionmail/BillionMail/issues/344) Sửa Mail Forward từ chối nhiều địa chỉ (mỗi dòng 1 email) | Bug | Trung bình | ✅ |
| 12 | [#373](https://github.com/Billionmail/BillionMail/issues/373) `open_rate`/`click_rate` biến mất khi delivered = 0 | Bug | Trung bình | ✅ |
| 13 | [#369](https://github.com/Billionmail/BillionMail/issues/369) `click_rate` dùng 2 mẫu số khác nhau trong cùng response | Bug | Trung bình | ✅ |
| 14 | [#368](https://github.com/Billionmail/BillionMail/issues/368) `is_sent = 3` không tồn tại; `ProcessTask` nuốt lỗi | Bug | Thấp | ✅ |
| 15 | [#372](https://github.com/Billionmail/BillionMail/issues/372) `RepairDKIMSigningConfig` bỏ qua lỗi & restart rspamd mỗi key | Bug | Trung bình | ✅ |
| 16 | [#370](https://github.com/Billionmail/BillionMail/issues/370) Lỗi tra relay âm thầm bật lại DKIM local; `GetRelayDomains` bỏ qua ctx | Bug | Thấp | ✅ |
| 17 | [#371](https://github.com/Billionmail/BillionMail/issues/371) Logger ctx `nil`, lỗi lookup unsubscribe bị nuốt, phụ thuộc grpc thừa | Code quality | Thấp | ✅ |
| 18 | [#382](https://github.com/Billionmail/BillionMail/issues/382) Kênh cảnh báo WebHook (Feishu, WeCom, DingTalk, Slack, Gotify…) | Tính năng | Trung bình | ✅ |
| 19 | [#215](https://github.com/Billionmail/BillionMail/issues/215) Header `X-Mailer` kèm phiên bản | Tính năng | Thấp | ✅ |
| 20 | [#202](https://github.com/Billionmail/BillionMail/issues/202) API gửi mail cho phép ghi đè subject | Tính năng | Trung bình | ✅ |

> Issue #280 (API tạo group trả về id), #191 (hiển thị dung lượng đã dùng/quota), #119 (tên ngôn ngữ
> theo bản ngữ) và #83 (truyền biến qua API) **đã có sẵn trong code của fork này**, nên được thay bằng #202.

---

## Chi tiết thay đổi

### 1. #390 – Token tracking giả mạo được + open redirect (bảo mật)
**Nguyên nhân:** token `/pmta/<token>` tự chứa luôn khoá AES và IV → ai cũng tạo được token hợp lệ để
ghi số liệu mở/click giả, hoặc biến domain mail thành open redirect.
**Sửa:**
- Token mới = `<payload mã hoá>.<HMAC-SHA256>` với secret do server giữ (`bm_options.pmta_tracker_secret`,
  tự sinh lần đầu) – `maillog_stat/tracker_signature.go`.
- `CampaignEventHandler` chỉ ghi thống kê / redirect khi chữ ký hợp lệ; chữ ký sai → 400.
- Token cũ (đã nằm trong mail gửi trước đây) vẫn mở được nhưng **không ghi thống kê và không tự redirect**:
  pixel trả ảnh trong suốt, click hiển thị trang xác nhận có link (đã escape).
- Chỉ redirect tới URL `http/https` tuyệt đối; chặn `javascript:`, `//evil`, `data:`…
- `Decrypt` không còn panic với ciphertext lỗi (độ dài không chia hết block).
- Test: `tracker_signature_test.go`.

### 2. SafePath chặn route công khai (#364, #381, #387, #338, #361, #360)
**Nguyên nhân:** hook `HookBeforeServe` yêu cầu session `safe_path_pass` cho mọi route động, kể cả
Roundcube (người dùng mailbox), ACME HTTP-01 (Let's Encrypt) và link tracking trong email.
**Sửa:** `internal/cmd/safepath.go` – danh sách route công khai: `/roundcube`, `/roundcube/*`,
`/.well-known/acme-challenge/*`, `/pmta/*`, `/landing/video`, `/inbox-placement/oauth/callback`.
Route quản trị và `/api/*` vẫn bị bảo vệ. Test: `safepath_test.go`.

### 3. #172 – Đọc file tùy ý
- Endpoint đã là stub; thay `status.Errorf` (grpc) bằng `gerror.NewCode(CodeNotSupported)`.
- Thay `validateFilePath` (chỉ chặn `./` và `/`) bằng `resolvePathWithin(base, path)`: chặn đường dẫn
  tuyệt đối, `..`, byte NUL, đảm bảo luôn nằm trong thư mục gốc. Test: `files/safe_path_test.go`.

### 4. #355 – Relay trộn credential giữa các domain
**Nguyên nhân:** `sasl_passwd` khoá theo `[host]:port`; nhiều relay cùng `smtp.resend.com:465` →
Postfix chỉ lấy 1 credential. Ngoài ra relay cùng tên hiển thị sinh trùng tên service trong `master.cf`.
**Sửa (`relay/config_sync.go`):**
- Bật `smtp_sender_dependent_authentication = yes`; ghi thêm dòng `@domain user:pass` cho mỗi domain
  được gán relay (vẫn giữ dòng `[host]:port` làm fallback).
- Block main.cf cũ được tự nâng cấp khi đồng bộ.
- Tên service custom thêm `_<id>` để không trùng.
- Test: `config_sync_test.go` (per-domain credentials, tên service, TLS theo cổng).

### 5. #379 – Test SMTP treo ở cổng 2465
**Sửa (`mail_service/sending.go`):**
- Timeout dial 15s, timeout handshake/greeting/AUTH 30s, timeout mỗi lần gửi 5 phút.
- Implicit TLS cho 465/2465/8465 (`IsImplicitTLSPort`), dùng chung cho `master.cf` của relay.
- STARTTLS dùng khi server quảng bá (không chỉ cổng 587); cổng khác (2525…) dùng `smtp_tls_security_level=may`.
- Báo lỗi rõ “port may require implicit TLS/SSL” khi server không chào.
- Bỏ qua AUTH khi không có credential (relay xác thực theo IP).
- Test: `sending_test.go` (server câm → lỗi trong < 5s).

### 6. #376 – Display Name bị bỏ qua
- **Nguyên nhân gốc:** `doSend` xoá header `RealName` khỏi map header dùng chung; khi kết nối SMTP
  giữ lâu bị đóng, `Send` reconnect rồi gọi lại `doSend` → lần gửi thứ 2 không còn display name nên
  From rơi về local-part (`vuanem <vuanem@…>`). Nay `doSend` làm việc trên bản sao header.
- Header From dựng bằng `net/mail.Address` (`FormatFromHeader`) → tên có dấu phẩy, ngoặc kép,
  tiếng Việt/Trung… được quote/encode đúng RFC 5322. Test: `sending_test.go`.

### 7. #389 – SPF sai IP khi multi-IP
`GetSPFRecordWithIP` + `FormatSPFValue`; `GetRecordsInCache` dùng IP riêng của domain cho SPF giống
bản ghi A (bỏ qua cache chỉ chứa IP chính). Test: `domains/spf_test.go`.

### 8. #378 – Import mailbox rỗng
- `AddImport` từ chối username rỗng/sai định dạng (`ValidateMailboxAddress`) và báo trùng thay vì đếm là thành công.
- Import nhận cột `email` như `username`, chuẩn hoá header (khoảng trắng, chữ hoa, BOM của Excel),
  ghép `local_part@domain` khi thiếu username.
- Mật khẩu mặc định khi import không còn là `123456` mà sinh ngẫu nhiên đúng chính sách.
- API xoá chấp nhận username rỗng để dọn bản ghi lỗi cũ.

### 9. #350 – Mật khẩu tạo hàng loạt
`GenerateCompliantPassword` (crypto/rand) luôn có chữ thường + chữ hoa + số, xáo trộn vị trí.
Test chạy 2000 lần. File: `mail_boxes/password.go`.

### 10. #348 – Switch trạng thái mailbox
Endpoint mới `POST /api/mailbox/set_status {username, active}` chỉ cập nhật `active`
(API update cũ bắt buộc `local_part` + password nên switch luôn lỗi, và có nguy cơ ghi đè mật khẩu).
Frontend `views/mailbox/index.vue` dùng endpoint mới.

### 11. #344 – Mail Forward nhiều địa chỉ
`ProcessForwardUsers` tách theo xuống dòng thật, `\n` escape, dấu phẩy, chấm phẩy; loại trùng.
Bỏ rule `v:"email"` trên toàn chuỗi, thay bằng kiểm tra từng địa chỉ (`InvalidForwardUsers`) ở cả add & edit.

### 12–13. #373, #369 – Tỷ lệ mở/click
- `getTaskMailProviders` luôn trả `open_rate`/`click_rate` (0 khi delivered = 0).
- Biểu đồ click rate dùng `clicked / delivered` giống dashboard; tỷ lệ theo provider ở overview dùng
  `delivered` làm mẫu số thay vì `sends`.

### 14. #368 – Task executor
`isTaskComplete` dùng `is_sent = 1` (có comment giải thích các trạng thái); `ProcessTask` trả lỗi
sau khi kiểm tra hoàn thành để caller phân biệt chạy lỗi một phần.

### 15–16. #372, #370 – DKIM
- Tách `ensureDKIMKeyPair`: chỉ sinh key (kiểm tra lại dưới mutex, tạo thư mục), không ghi config/restart.
- `RepairDKIMSigningConfig` báo lỗi và **bỏ qua domain** không sinh được key (không trỏ rspamd vào file
  không tồn tại), tạo thư mục `dkim` nếu thiếu, chỉ restart rspamd 1 lần.
- Lỗi tra relay không còn âm thầm ký DKIM local (tránh DKIM trùng #289) – để repair xử lý.
- `GetRelayDomains` dùng `ctx`.

### 17. #371 – Dọn dẹp
Logger dùng `context.Background()` thay `nil`; log lỗi lookup sender trong unsubscribe;
loại `google.golang.org/grpc` khỏi `go.mod` (`go mod tidy`).

### 18. #382 – WebHook cảnh báo
- Backend: `domains/alert_webhook.go`, API `GET /api/settings/get_alert_webhook`,
  `POST /api/settings/set_alert_webhook` (có tuỳ chọn gửi thử). Định dạng: JSON, Feishu/Lark, WeCom,
  DingTalk, Slack, Gotify. Hoạt động độc lập với cấu hình SMTP cảnh báo.
- Frontend: nút “Webhook” trong Cài đặt → Thông báo → Blacklist (`WebhookSettings.vue`), i18n en/zh/ja.
- Test: `alert_webhook_test.go` (httptest).

### 19. #215 – X-Mailer
`X-Mailer: BillionMail <version>` lấy từ `server.version`.

### 20. #202 – API ghi đè subject
Trường `subject` (tuỳ chọn) cho `/api/batch_mail/api/send` và `/api/batch_mail/api/batch_send`,
lưu vào cột mới `api_mail_logs.subject` (tự migrate), hỗ trợ biến template như subject gốc.

---

## Phần 2 – Đánh giá các issue còn lại & làm thêm

Sau 20 mục trên, tôi đọc lại toàn bộ các issue còn lại và chia làm 4 nhóm: **(A) làm thêm ngay**,
**(B) đã có sẵn trong fork / đã được Phần 1 xử lý**, **(C) hỏi đáp – cấu hình – môi trường (không cần sửa code)**,
**(D) tính năng lớn – để lại** (kèm lý do).

### A. Đã làm thêm (code + test)

| # | Issue | Nội dung thay đổi |
|---|-------|-------------------|
| 21 | [#367](https://github.com/Billionmail/BillionMail/issues/367) Thêm tiếng Việt | `frontend/src/i18n/lang/vi.json` dịch **100% key** (test so khớp key với en), locale Naive UI `viVN`; backend `languages/vi/server.json` dịch **348/348** thông báo; thêm "Tiếng Việt" vào `languages/settings.json`. Sửa luôn lỗi chọn ngôn ngữ `tr` làm hỏng `bt-config-provider` (fallback về en). |
| 22 | [#86](https://github.com/Billionmail/BillionMail/issues/86) / [#357](https://github.com/Billionmail/BillionMail/issues/357) Reply-To | Trường `reply_to` cho chiến dịch và API template (cột mới tự migrate), header `Reply-To` được kiểm tra/encode (`FormatReplyTo`, chặn header injection). Form chiến dịch & form API có ô nhập + validate. |
| 23 | [#362](https://github.com/Billionmail/BillionMail/issues/362) / [#343](https://github.com/Billionmail/BillionMail/issues/343) Gửi dồn một lúc, cần delay giữa các email | Trường `send_interval` (0–3600 giây) cho chiến dịch; `intervalPacer` đảm bảo khoảng cách tối thiểu giữa 2 email, độc lập với rate controller theo luồng (vốn cho burst ≥ 10 và ~20 mail/giây/luồng). UI: ô "Khoảng cách gửi (giây)". |
| 24 | [#331](https://github.com/Billionmail/BillionMail/issues/331) Plain-text / multipart | Email HTML được gửi dạng `multipart/alternative` với phần `text/plain` sinh tự động (`HTMLToText`: bỏ style/script/pixel, giữ link dạng `text (url)`, list, xuống dòng) → tránh điểm phạt `MIME_HTML_ONLY`. |
| 25 | [#303](https://github.com/Billionmail/BillionMail/issues/303) Tắt nhận thư theo domain (domain chỉ gửi) | Cột `domain.receive_enabled`, API `POST /api/domains/set_receive`, công tắc "Nhận thư" trong danh sách domain. Khi tắt, Postfix không coi domain là local → thư gửi tới domain đó đi theo MX thay vì lỗi "User unknown in virtual mailbox table". Câu query Postfix được tự vá khi khởi động (idempotent, giữ CRLF). Domain mới mặc định vẫn nhận thư. |
| 26 | [#153](https://github.com/Billionmail/BillionMail/issues/153) / [#212](https://github.com/Billionmail/BillionMail/issues/212) CSV import bị lỗi font | Frontend đọc file bằng `readTextFile`: UTF-8 (bỏ BOM) → GB18030 → Windows-1252; áp dụng cho import liên hệ, nhóm, blacklist, mailbox. Backend bỏ BOM ở CSV liên hệ (BOM làm cột `email` không nhận ra). |

### B. Đã có sẵn trong fork hoặc được Phần 1 xử lý (không cần làm thêm)

| Issue | Tình trạng |
|-------|-----------|
| #280 API tạo group trả id · #191 hiển thị đã dùng/quota · #119 tên ngôn ngữ bản ngữ · #83 truyền biến qua API · #96 spintax · #93 trục Y tự co giãn | Đã có trong code |
| #78 catch-all | Có catch-all theo domain (Domain → Catch All) và alias `@domain` |
| #208 / #353 SSL tự upload | Có `SetSSL` (dán KEY/CRT) |
| #115 / #103 thêm/nhập mailbox hàng loạt · #136 lọc mailbox theo domain | Đã có |
| #183 / #288 tự xoá thư cũ | Có "Thời gian lưu thư" (thùng rác) trong Cài đặt |
| #113 / #54 link unsubscribe/tracking sai domain sau reverse proxy | Có cài đặt Reverse Proxy domain |
| #55 relay không xác thực | Có `AuthMethod = NONE`; Phần 1 bỏ AUTH khi không có user/pass |
| #337 cảnh báo `BM_API_KEY` | Fork này không dùng biến đó |
| #345, #321, #174, #132, #81, #59, #149 thống kê mở/click = 0 | Nguyên nhân chính là SafePath chặn `/pmta/*` → đã sửa ở mục 2; mẫu số tỷ lệ sửa ở #369/#373 |
| #160, #346, #108, #89, #30 lỗi xin SSL Let's Encrypt | ACME challenge bị SafePath chặn → đã sửa ở mục 2 (các lỗi rate-limit/DNS còn lại là cấu hình) |
| #167, #156, #177, #171, #101, #102, #169 relay báo `530/550 Authentication required` | Rất có thể do credential bị trộn giữa các relay cùng host → đã sửa ở #355 |
| #344 + #330 forward nhiều địa chỉ | Phần 1 sửa (nhiều đích, phân tách bằng dòng/dấu phẩy) |

### C. Hỏi đáp / cấu hình / môi trường – không sửa code

DNS – bản ghi A/MX/SPF/DKIM chưa nhận, Cloudflare proxy, PTR: #389 (đã sửa), #341, #329, #284, #213, #211, #205, #168, #165, #137, #110, #107, #105, #94, #91, #80, #71, #68, #67, #65, #62, #45, #44, #40, #39.
Cổng 25 bị chặn / không gửi được / spam: #356, #312, #290, #262, #261, #241, #218, #166, #162, #133, #79, #70, #69, #51, #20, #14.
Cài đặt Docker/WSL/ARM/macOS/aaPanel/cổng bận: #386, #302, #299, #291, #187, #184, #148, #146, #135, #124, #77, #47, #32, #28, #26, #13, #6, #207, #195.
Truy cập giao diện / Not Found do SafePath / reverse proxy: #279, #305, #273, #272, #253, #164, #61, #35, #16, #283, #210.
Khác (hỏi cách dùng): #349, #326, #324, #319, #310, #308, #304, #301, #259, #252, #251, #234, #231, #230, #222, #217, #214, #196, #158, #157, #147, #144, #141, #140, #129, #125, #118, #99, #98, #92, #72, #57, #56, #50, #38, #33, #31, #27, #5, #316 (tracker), #277 (hỏi về báo cáo bảo mật).
Spam/quảng cáo: #388 (DemonRoute) – không đưa vào.

### D. Tính năng lớn – để lại cho đợt sau

| Issue | Lý do chưa làm |
|-------|---------------|
| #325, #142 analytics theo từng email, lọc/xuất theo trạng thái | Cần thiết kế bảng/truy vấn mới + UI lớn |
| #307, #123, #176, #21 nhiều người dùng / nhóm / phân quyền | Thay đổi mô hình RBAC toàn hệ thống |
| #336 admin xem hộp thư người dùng · #340 tự đăng ký tài khoản | Nhạy cảm bảo mật, cần quyết định sản phẩm |
| #267, #179 xem toàn bộ inbox trong admin · #180 auto-reply · #150, #143 webhook thư đến/sự kiện | Cần tích hợp Dovecot/Sieve |
| #266, #19 HA đa máy, DB ngoài | Thay đổi kiến trúc triển khai |
| #109, #138, #29, #17, #85, #24 API CRUD đầy đủ, giới hạn gửi theo domain | Phạm vi lớn |
| #95 automation builder · #122 realtime GUI · #270 thiết kế lại UI | Phạm vi lớn |
| #127 SES configuration set · #204 Office 365 API · #209 proxy SOCKS · #185 gửi không cần domain | Tích hợp bên thứ ba |
| #365 preset Werkmail · #60 preset Resend | Chưa xác minh được thông số SMTP chính thức |
| #193, #97 thêm nhà cung cấp mail tuỳ chỉnh · #75 điểm spam rspamd trong UI · #76 đổi đăng nhập không cần restart · #82 xoá liên hệ theo điều kiện · #252 import 3000 mailbox gây quá tải · #269 hostname theo domain · #178 xuất mật khẩu dạng rõ (mâu thuẫn bảo mật với #27) · #73 migrate từ Mailcow · #100, #286 Railway/Easypanel | Ưu tiên thấp hơn / cần thảo luận |

---

## Phần 3 – Tiếng Việt cho toàn bộ giao diện (frontend)

**Đã có từ trước (#367):** `vi.json` đủ key so với `en.json`, mục "Tiếng Việt" trong `languages/settings.json`, ngôn ngữ của Naive UI (`viVN`), thông báo backend `languages/vi/server.json`.

**Làm thêm đợt này:** chuyển khoảng 140 chuỗi tiếng Anh viết cứng sang key i18n và dịch đủ 4 ngôn ngữ en/zh/ja/vi.

| Khu vực | File chính | Nội dung |
|---|---|---|
| Trình soạn email (kéo thả) | `features/EmailEditor/components/**` (khoảng 40 file) | Tên khối, tiêu đề panel, nhãn thuộc tính (màu, viền, padding, căn lề…), ô gợi ý thả nội dung. Nhãn trong `useNormalForm` được dịch lúc render. Key mới: `editor.*` |
| Domain | `views/domain/**` | Cột Not in Spam List / Current usage / A record, Hostname, trang Typography, câu "Checking … against … blacklists" (dùng `<i18n-t>`) |
| Chiến dịch & template | `market/task/components/*`, `market/template/*`, `template/components/*` | Cột bảng TaskStatus / TaskStatusDetails, trạng thái Confirmed/Unconfirmed, Name/Type/Drag |
| Contacts | `contacts/group/*`, `contacts/settings`, `contacts/suspend/Scan.vue` | Add group, List Type, Single/Double Opt-in, tab Unsubscribe, nút Scan |
| Mailbox | `MailboxForm`, `MailboxBatchAdd`, `MailboxExport` | Display Name, Quota limit, Format |
| Settings | `ai-model/*`, `SSLSettings.vue` | Thêm model/provider, Cancel/Confirm, hộp thoại xoá, nội dung hướng dẫn xin chứng chỉ SSL |
| Khác | `video-outreach`, `api/ApiTest`, `ApiKey`, `smtp/SmtpForm`, `bt-table-password`, `bt-table-page` ("Total"), `api/modules/public.ts` (thông báo tiếng Trung khi đổi ngôn ngữ) | |

**Cố ý giữ nguyên:** tên riêng và thuật ngữ kỹ thuật (BillionMail, Rspamd, Spintax, SPF/DKIM/DMARC/PTR, TXT, CSV, HTML, BaseURL, ApiKey, LOGIN/PLAIN/CRAM-MD5, Gmail/SendGrid…), nội dung mặc định chèn vào email (Button Text, Unsubscribe, Powered by), log console cho dev.

**Test mới:** `src/i18n/coverage.test.ts` kiểm tra rằng zh/ja/vi giữ đúng placeholder `{x}` của en, vi không có bản dịch rỗng, và mọi key `editor.*`/`ui.*` dùng trong code đều có trong en.json.

**Ngôn ngữ mặc định: Tiếng Việt.** Backend `public.DefaultLanguage = "vi"` (`internal/service/public/common.go`) được dùng khi admin chưa chọn ngôn ngữ; bỏ bước tự lấy ngôn ngữ trình duyệt (Accept-Language) vốn tự lưu `en` vào DB. Frontend: `i18n` locale và `globalStore.lang` mặc định `vi`; fallback vẫn là `en`. Cài đặt cũ đã có `bm_options.CurrentLanguage` thì giữ nguyên ngôn ngữ đã lưu.

---

## Kiểm tra
- `GOOS=linux go build ./...` và `go vet ./...` – sạch; `gofmt` các file đã sửa – sạch.
- `go test -count=1 -short ./internal/service/... ./internal/controller/... ./internal/cmd/` (chạy trong WSL Linux) – **tất cả pass**.
- Frontend: `eslint . --ext .vue,.ts,.tsx` – sạch; `vitest run` – **59 file / 447 test pass**; `rsbuild build` – thành công.
- Chưa chạy end-to-end với Docker (Postfix/Dovecot/Rspamd thật) – nên thử trên staging các mục đụng tới Postfix: #355 (sasl theo domain), #303 (receive filter), #331 (multipart).
