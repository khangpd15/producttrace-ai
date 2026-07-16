# T├ái liß╗çu B├án giao: Module Qu├¬n/─Éß╗òi Mß║¡t khß║⌐u (Fullstack) & SendGrid

## Tß╗òng quan
Dß╗▒ ├ín ─æ├ú ─æ╞░ß╗úc t├¡ch hß╗úp ho├án chß╗ënh luß╗ông **Qu├¬n Mß║¡t Khß║⌐u (Forgot Password)** tß╗½ Backend (NestJS) tß╗¢i Frontend (React/Vite). To├án bß╗Ö hß╗ç thß╗æng gß╗¡i email qua SendGrid v├á x├íc thß╗▒c mß║¡t khß║⌐u ─æß╗üu ─æ├ú hoß║ít ─æß╗Öng tr╞ín tru. Backend hiß╗çn ─æang sß╗¡ dß╗Ñng c╞í sß╗ƒ dß╗» liß╗çu giß║ú lß║¡p (Mock JSON) vß╗¢i thiß║┐t kß║┐ chuß║⌐n Clean Architecture ─æß╗â dß╗à d├áng thay ─æß╗òi sang Prisma/PostgreSQL sau n├áy.

---

## 1. Backend (NestJS - `nest-ai-service`)

### Kiß║┐n tr├║c & Database (Mock JSON)
- Kh├┤ng sß╗¡ dß╗Ñng CSDL thß║¡t ─æß╗â dß╗à d├áng test. Dß╗» liß╗çu ─æ╞░ß╗úc l╞░u trß╗▒c tiß║┐p v├áo 2 file JSON ─æ├│ng vai tr├▓ nh╞░ database:
  - `src/mock-data/users.json` (L╞░u th├┤ng tin User v├á Mß║¡t khß║⌐u b─âm - `bcrypt`)
  - `src/mock-data/password-reset-tokens.json` (L╞░u Token qu├¬n mß║¡t khß║⌐u)
- ├üp dß╗Ñng **Repository Pattern**: `AuthService` ho├án to├án giao tiß║┐p qua c├íc Interface (`UserRepository`, `PasswordResetRepository`). Khi t├¡ch hß╗úp database thß║¡t, Dev chß╗ë cß║ºn viß║┐t class `PrismaUserRepository` v├á thay ─æß╗òi DI provider trong `AuthModule`. Kh├┤ng cß║ºn sß╗¡a ─æß╗òi logic nghiß╗çp vß╗Ñ.

### C├íc API ─æ├ú ho├án thiß╗çn
1. **`POST /auth/forgot-password`**
   - **Body**: `{ "email": "test@gmail.com" }`
   - Tß║ío token bß║úo mß║¡t bß║▒ng `crypto.randomBytes()`.
   - L╞░u trß╗» m├ú b─âm (hash) cß╗ºa token (chuß║⌐n bß║úo mß║¡t), c├│ hß║ín 15 ph├║t.
   - Gß╗ìi `MailService` ─æß╗â gß╗¡i SendGrid email (c├│ chß╗⌐a link tß╗¢i Frontend).
   - *L╞░u ├╜*: API lu├┤n trß║ú vß╗ü `200 OK` kß╗â cß║ú email kh├┤ng tß╗ôn tß║íi ─æß╗â chß╗æng lß║íi lß╗ù hß╗òng d├▓ qu├⌐t email (Email Enumeration).

2. **`GET /auth/validate-reset-token`**
   - **Query**: `?token=...&email=...`
   - Kiß╗âm tra t├¡nh hß╗úp lß╗ç v├á thß╗¥i hß║ín cß╗ºa token. API n├áy ─æ╞░ß╗úc Frontend tß╗▒ ─æß╗Öng gß╗ìi khi User mß╗ƒ link trong email.

3. **`POST /auth/reset-password`**
   - **Query**: `?email=...` | **Body**: `{ "token": "...", "password": "...", "confirmPassword": "..." }`
   - Kiß╗âm tra token mß╗Öt lß║ºn nß╗»a.
   - M├ú ho├í (hash) mß║¡t khß║⌐u mß╗¢i bß║▒ng `bcrypt`.
   - Update mß║¡t khß║⌐u trong JSON v├á xo├í token ─æ├│ ─æi ─æß╗â ─æß║úm bß║úo Link chß╗ë d├╣ng ─æ╞░ß╗úc 1 lß║ºn duy nhß║Ñt.

---

## 2. Frontend (React + Vite + TailwindCSS - `web-frontend`)

### Giao diß╗çn & Luß╗ông xß╗¡ l├╜
- ─É├ú kh├┤i phß╗Ñc ho├án chß╗ënh cß║Ñu tr├║c Vite/React bß╗ï thiß║┐u tr╞░ß╗¢c ─æ├│ (`package.json`, `vite.config.ts`, `main.tsx`, `App.tsx`, c├ái ─æß║╖t TailwindCSS...).
- ─É├ú thiß║┐t kß║┐ 3 trang ch├¡nh d├╣ng React Router:
  1. `/login`: Giao diß╗çn ─æ─âng nhß║¡p.
  2. `/forgot-password`: M├án h├¼nh ─æiß╗ün email ─æß╗â y├¬u cß║ºu ─æß╗òi mß║¡t khß║⌐u.
  3. `/reset-password`: M├án h├¼nh thay ─æß╗òi mß║¡t khß║⌐u. ─É├ú xß╗¡ l├╜ kß╗╣ l╞░ß╗íng UX/UI:
     - Tß╗▒ ─æß╗Öng gß╗ìi API kiß╗âm tra Token (`cache: 'no-store'`) ─æß╗â tr├ính lß╗ùi cache tr├¼nh duyß╗çt khi click link nhiß╗üu lß║ºn.
     - Kiß╗âm tra sß╗⌐c mß║ính mß║¡t khß║⌐u.
     - Hiß╗ân thß╗ï m├án h├¼nh Lß╗ùi/Th├ánh c├┤ng r├╡ r├áng, tr├ính hiß╗ân thß╗ï form nß║┐u Token kh├┤ng hß╗úp lß╗ç.

### C├ích chß║íy Frontend
```bash
cd apps/web-frontend
npm install
npm run dev
```
Trang web sß║╜ chß║íy ß╗ƒ `http://localhost:5173`.

---

## 3. Cß║Ñu h├¼nh M├┤i tr╞░ß╗¥ng (.env)

─Éß║úm bß║úo file `.env` ß╗ƒ th╞░ mß╗Ñc root c├│ ─æß║ºy ─æß╗º c├íc biß║┐n sau:

```env
# URL cß╗ºa Frontend ─æß╗â NestJS nh├║ng v├áo nß╗Öi dung Email SendGrid
FRONTEND_URL=http://localhost:5173

# Cß║Ñu h├¼nh SendGrid
SENDGRID_API_KEY=m├ú_api_sendgrid_cß╗ºa_bß║ín_ß╗ƒ_─æ├óy
FROM_EMAIL=email_ng╞░ß╗¥i_gß╗¡i_─æ├ú_─æ╞░ß╗úc_x├íc_thß╗▒c_ß╗ƒ_─æ├óy
WELCOME_TEMPLATE_ID=id_template_ch├áo_mß╗½ng_ß╗ƒ_─æ├óy
RESET_PASSWORD_TEMPLATE_ID=id_template_qu├¬n_mß║¡t_khß║⌐u_ß╗ƒ_─æ├óy
```

**Ch├║ ├╜ vß╗ü Template HTML cß╗ºa SendGrid:**
- File template HTML d├ánh cho SendGrid nß║▒m tß║íi `reset-password-email.html` (─æ├ú ─æ╞░ß╗úc Dev cung cß║Ñp vß╗¢i thiß║┐t kß║┐ bß║úng t╞░╞íng th├¡ch tß╗æt tr├¬n mß╗ìi nß╗ün tß║úng).
- Template c├│ c├íc biß║┐n: `{{name}}`, `{{resetLink}}`, `{{year}}`, `{{companyName}}`. 
- NestJS ─æ├ú ─æ╞░ß╗úc cß║Ñu h├¼nh truyß╗ün ─æß╗º dß╗» liß╗çu cho c├íc biß║┐n n├áy. Ri├¬ng `{{companyName}}` bß║ín c├│ thß╗â hardcode thß║│ng v├áo cß║Ñu h├¼nh SendGrid.

## 4. H╞░ß╗¢ng dß║½n Test (Cho Tester/Dev kh├íc)
1. ─Éß║úm bß║úo chß║íy cß║ú NestJS (`npm run start:dev`) v├á Vite Frontend (`npm run dev`).
2. Mß╗ƒ file `apps/nest-ai-service/src/mock-data/users.json` v├á ─æiß╗ün email thß║¡t cß╗ºa bß║ín v├áo (─æß╗â nhß║¡n ─æ╞░ß╗úc th╞░).
3. Mß╗ƒ tr├¼nh duyß╗çt v├áo `http://localhost:5173/login`, bß║Ñm "Forgot your password?".
4. Nhß║¡p email v├á submit.
5. Check hß╗Öp th╞░ ─æß║┐n, mß╗ƒ email v├á click v├áo n├║t "X├íc nhß║¡n thay ─æß╗òi mß║¡t khß║⌐u".
6. ─Éiß╗ün mß║¡t khß║⌐u mß╗¢i tr├¬n form hiß╗çn ra.
7. Reset th├ánh c├┤ng -> Kiß╗âm tra file `users.json` sß║╜ thß║Ñy chuß╗ùi b─âm `password` ─æ├ú bß╗ï thay ─æß╗òi, v├á `password-reset-tokens.json` ─æ├ú rß╗ùng. Mß╗ìi thß╗⌐ hoß║ít ─æß╗Öng nh╞░ mß╗Öt hß╗ç thß╗æng thß╗▒c thß╗Ñ!

---


# T├ái liß╗çu B├án giao: Module Th├┤ng B├ío Bß║úo H├ánh (Warranty Notification)

## Tß╗òng quan

─É├ú triß╗ân khai ho├án chß╗ënh t├¡nh n─âng gß╗¡i **email th├┤ng b├ío cß║¡p nhß║¡t trß║íng th├íi bß║úo h├ánh** cho kh├ích h├áng th├┤ng qua hß╗ç thß╗æng Event-Driven (RabbitMQ ΓåÆ NestJS Worker ΓåÆ SendGrid).

- **Use Case**: `UC-P3-NOTI-01` ΓÇö Gß╗¡i th├┤ng b├ío cß║¡p nhß║¡t bß║úo h├ánh tß╗¢i email kh├ích h├áng khi trß║íng th├íi thay ─æß╗òi.
- **Nh├ính Git**: `feature/notification-warranty`
- **PR**: [#90 ΓåÆ develop](https://github.com/khangpd15/producttrace-ai/pull/90) _(─æang chß╗¥ review & approve)_
- **SendGrid Template ID**: `d-aa9b56ba4bf64b54a72eddc7ba33ba03`

---

## 1. Ph├ón chia tr├ích nhiß╗çm (Architecture: Event-Driven ΓÇö Ph╞░╞íng ├ín A)

| Service | Tr├ích nhiß╗çm |
|---|---|
| **Go Core Service** | Truy vß║Ñn PostgreSQL lß║Ñy th├┤ng tin bß║úo h├ánh thß║¡t ΓåÆ Publish event `notification.sent` l├¬n RabbitMQ vß╗¢i ─æß║ºy ─æß╗º payload |
| **NestJS (`nest-ai-service`)** | Lß║»ng nghe event tß╗½ RabbitMQ ΓåÆ Gß╗¡i email qua SendGrid ΓÇö **─É├ú ho├án th├ánh** Γ£à |

> **NestJS kh├┤ng kß║┐t nß╗æi trß╗▒c tiß║┐p v├áo database.** To├án bß╗Ö dß╗» liß╗çu bß║úo h├ánh (t├¬n sß║ún phß║⌐m, trß║íng th├íi, ng├áy hß║┐t hß║ín...) ─æ╞░ß╗úc lß║Ñy tß╗½ Go Core Service th├┤ng qua payload RabbitMQ.

---

## 2. Luß╗ông hoß║ít ─æß╗Öng (Event Flow)

```
[PostgreSQL] ΓåÉ Go Core Service truy vß║Ñn bß║úng warranties + users + products
     Γöé
     Γöé  Publish RabbitMQ event:
     Γöé  Exchange: "product-trace.events"
     Γöé  Routing key: "notification.sent"
     Γöé  Payload: { event_type, data: { email, full_name, product_name, warranty_status, warranty_end_date } }
     Γû╝
RabbitMQ Queue: "ai.events"
     Γöé
     Γû╝
NotificationConsumer (NestJS Worker)  ΓåÉ lß║»ng nghe li├¬n tß╗Ñc
     Γöé  case "notification.sent"
     Γöé  Tr├¡ch xuß║Ñt payload ─æß╗Öng
     Γû╝
MailService.sendWarrantyUpdateEmail()
     Γöé  Gß╗ìi SendGrid API vß╗¢i templateId + dynamicTemplateData
     Γû╝
Email ─æß║┐n h├▓m th╞░ kh├ích h├áng  Γ£à
```

---

## 3. C├íc file ─æ├ú thay ─æß╗òi trong `nest-ai-service`

| File | M├┤ tß║ú thay ─æß╗òi |
|---|---|
| `src/messaging/rabbitmq/rabbitmq.constants.ts` | Th├¬m `NOTIFICATION_SENT: 'notification.sent'` v├áo `ROUTING_KEYS` v├á `EVENT_TYPES`; merge th├¬m c├íc keys mß╗¢i tß╗½ nh├ính develop (`OWNERSHIP_OTP`, `TRACE_*`, `EMBEDDING_*`) |
| `src/messaging/rabbitmq/rabbitmq.service.ts` | Th├¬m `NOTIFICATION_SENT` v├á `OWNERSHIP_OTP` v├áo danh s├ích `routingKeys` ─æß╗â tß╗▒ ─æß╗Öng bind h├áng ─æß╗úi khi khß╗ƒi tß║ío |
| `src/messaging/consumers/notification.consumer.ts` | Th├¬m 3 tr╞░ß╗¥ng ─æß╗Öng v├áo `NotificationPayload` (`product_name`, `warranty_status`, `warranty_end_date`); th├¬m `case "notification.sent"` trong switch ─æß╗â gß╗ìi `sendWarrantyUpdateEmail` |
| `src/mail/mail.service.ts` | Method `sendWarrantyUpdateEmail(to, fullName, productName, status, endDate)` ΓÇö gß╗¡i qua SendGrid Dynamic Template hoß║╖c fallback HTML thuß║ºn |

---

## 4. Cß║Ñu tr├║c Payload RabbitMQ

Go Core Service **bß║»t buß╗Öc** publish ─æ├║ng schema sau:

```json
{
  "event_type": "notification.sent",
  "data": {
    "email": "khachhang@gmail.com",
    "full_name": "Nguyß╗àn V─ân A",
    "product_name": "iPhone 15 Pro Max 256GB",
    "warranty_status": "─É├ú ho├án tß║Ñt sß╗¡a chß╗»a",
    "warranty_end_date": "24/10/2026"
  }
}
```

> **L╞░u ├╜**: Tß║Ñt cß║ú 5 tr╞░ß╗¥ng ─æß╗üu l├á dß╗» liß╗çu thß║¡t (dynamic). Nß║┐u thiß║┐u tr╞░ß╗¥ng n├áo, hß╗ç thß╗æng c├│ gi├í trß╗ï fallback mß║╖c ─æß╗ïnh ─æß╗â tr├ính crash ΓÇö nh╞░ng email sß║╜ thiß║┐u th├┤ng tin.

---

## 5. Cß║Ñu h├¼nh SendGrid Template

**Template ID**: `d-aa9b56ba4bf64b54a72eddc7ba33ba03`

| T├¬n biß║┐n SendGrid | Dß╗» liß╗çu truyß╗ün v├áo | V├¡ dß╗Ñ |
|---|---|---|
| `{{fullName}}` | T├¬n kh├ích h├áng | `Nguyß╗àn V─ân A` |
| `{{productName}}` | T├¬n sß║ún phß║⌐m | `iPhone 15 Pro Max 256GB` |
| `{{status}}` | Trß║íng th├íi bß║úo h├ánh | `─É├ú ho├án tß║Ñt sß╗¡a chß╗»a` |
| `{{endDate}}` | Ng├áy hß║┐t hß║ín bß║úo h├ánh | `24/10/2026` |
| `{{frontendUrl}}` | Link hß╗ç thß╗æng | `http://localhost:5173` |
| `{{year}}` | N─âm hiß╗çn tß║íi (auto) | `2026` |

---

## 6. Cß║Ñu h├¼nh m├┤i tr╞░ß╗¥ng (.env)

Th├¬m biß║┐n sau v├áo file `.env` cß╗ºa `apps/nest-ai-service`:

```env
# SendGrid
SENDGRID_API_KEY=your_api_key
SENDGRID_FROM_EMAIL=khangpd.ce191105@gmail.com

# Template ID cho th├┤ng b├ío cß║¡p nhß║¡t bß║úo h├ánh
WARRANTY_UPDATE_TEMPLATE_ID=d-aa9b56ba4bf64b54a72eddc7ba33ba03
```

> Nß║┐u `WARRANTY_UPDATE_TEMPLATE_ID` kh├┤ng ─æ╞░ß╗úc set, hß╗ç thß╗æng d├╣ng Template ID tr├¬n l├ám mß║╖c ─æß╗ïnh.
> Nß║┐u `SENDGRID_API_KEY` kh├┤ng ─æ╞░ß╗úc set, hß╗ç thß╗æng chß║íy ß╗ƒ **MOCK mode** ΓÇö chß╗ë log ra console.

---

## 7. H╞░ß╗¢ng dß║½n Test (Tester / Dev kh├íc)

### Test thß╗º c├┤ng qua RabbitMQ Management UI
1. Mß╗ƒ `http://localhost:15672` (─æ─âng nhß║¡p: `admin` / `admin123`).
2. V├áo **Exchanges** ΓåÆ chß╗ìn exchange `product-trace.events`.
3. T├¼m mß╗Ñc **Publish message**, ─æiß╗ün:
   - **Routing key**: `notification.sent`
   - **Payload**:
     ```json
     {
       "event_type": "notification.sent",
       "data": {
         "email": "test@gmail.com",
         "full_name": "Nguyß╗àn V─ân A",
         "product_name": "iPhone 15 Pro Max",
         "warranty_status": "ACTIVE",
         "warranty_end_date": "24/10/2026"
       }
     }
     ```
4. Nhß║Ñn **Publish message**.
5. Kiß╗âm tra log cß╗ºa `nest-ai-service`:
   ```
   [NotificationConsumer] Warranty update email sent to test@gmail.com
   ```
6. Kiß╗âm tra h├▓m th╞░ nhß║¡n email vß╗¢i giao diß╗çn tß╗½ SendGrid template.

---

## 8. ─Éiß╗âm quan trß╗ìng cho Dev tiß║┐p nhß║¡n

### Dev Go Core Service cß║ºn l├ám:
- Sau khi cß║¡p nhß║¡t trß║íng th├íi bß║úo h├ánh trong PostgreSQL, publish event l├¬n RabbitMQ vß╗¢i:
  - **Exchange**: `product-trace.events`
  - **Routing key**: `notification.sent`
  - **Payload**: JSON ─æ├║ng schema mß╗Ñc 4 ß╗ƒ tr├¬n
- Tham khß║úo c├íc routing key ─æ├ú c├│ trong `apps/go-core-service/internal/events/rabbitmq/constants.go`

### Dev NestJS KH├öNG cß║ºn sß╗¡a g├¼ th├¬m khi:
- Go Core Service thay ─æß╗òi nß╗Öi dung bß║úo h├ánh ΓÇö chß╗ë cß║ºn ─æß║úm bß║úo payload JSON ─æ├║ng schema.

### ─Éß╗â th├¬m loß║íi th├┤ng b├ío mß╗¢i (v├¡ dß╗Ñ: `warranty.expired`):
1. Th├¬m constant mß╗¢i v├áo `rabbitmq.constants.ts`
2. Bind routing key mß╗¢i v├áo `rabbitmq.service.ts`
3. Th├¬m `case` mß╗¢i v├áo `notification.consumer.ts`
4. Th├¬m method mß╗¢i v├áo `mail.service.ts`

### Template fallback:
Nß║┐u ch╞░a setup SendGrid Template, hß╗ç thß╗æng tß╗▒ ─æß╗Öng gß╗¡i email HTML thuß║ºn c├│ ─æß║ºy ─æß╗º th├┤ng tin (kh├┤ng bß╗ï lß╗ùi/crash).

---

## 9. Lß╗ïch sß╗¡ thay ─æß╗òi

| Ng├áy | Nß╗Öi dung |
|---|---|
| 14/07/2026 | Khß╗ƒi tß║ío module, implement `sendWarrantyUpdateEmail` vß╗¢i 5 tham sß╗æ ─æß╗Öng |
| 14/07/2026 | Bind routing key `notification.sent` v├áo `rabbitmq.service.ts` |
| 14/07/2026 | Giß║úi quyß║┐t merge conflict vß╗¢i nh├ính `develop` (th├¬m `OWNERSHIP_OTP`, `TRACE_*`, `EMBEDDING_*` keys) |
| 14/07/2026 | Khß╗ƒi tß║ío module, implement `sendWarrantyUpdateEmail` vß╗¢i 5 tham sß╗æ ─æß╗Öng |
| 14/07/2026 | Bind routing key `notification.sent` v├áo `rabbitmq.service.ts` |
| 14/07/2026 | Giß║úi quyß║┐t merge conflict vß╗¢i nh├ính `develop` (th├¬m `OWNERSHIP_OTP`, `TRACE_*`, `EMBEDDING_*` keys) |
| 14/07/2026 | Tß║ío PR #90 ΓåÆ `develop`, ─æang chß╗¥ review |

---

# T├ái liß╗çu B├án giao: Module Th├┤ng B├ío Chuyß╗ân Quyß╗ün Sß╗ƒ Hß╗»u (Ownership Notification Worker)

## Tß╗òng quan

─É├ú triß╗ân khai ho├án chß╗ënh t├¡nh n─âng gß╗¡i **email th├┤ng b├ío chuyß╗ân quyß╗ün sß╗ƒ hß╗»u sß║ún phß║⌐m** cho kh├ích h├áng th├┤ng qua hß╗ç thß╗æng Event-Driven (RabbitMQ ΓåÆ NestJS Worker ΓåÆ SendGrid).

- **Use Case**: Gß╗¡i email th├┤ng b├ío cho ng╞░ß╗¥i nhß║¡n khi quyß╗ün sß╗ƒ hß╗»u sß║ún phß║⌐m ─æ╞░ß╗úc chuyß╗ân giao th├ánh c├┤ng.
- **Nh├ính Git**: `feature/notification-ownership`
- **PR**: [Mß╗ƒ PR ΓåÆ develop](https://github.com/khangpd15/producttrace-ai/compare/develop...feature/notification-ownership?expand=1) _(─æang chß╗¥ review & approve)_
- **SendGrid Template ID**: `d-1f78adcf1e3644e6bee66c2ea402af69`

---

## 1. Ph├ón chia tr├ích nhiß╗çm (Architecture: Event-Driven ΓÇö Queue chung)

| Service | Tr├ích nhiß╗çm |
|---|---|
| **Go Core Service** | Sau khi chuyß╗ân quyß╗ün sß╗ƒ hß╗»u th├ánh c├┤ng trong PostgreSQL ΓåÆ Publish event `ownership.transferred` l├¬n RabbitMQ vß╗¢i ─æß║ºy ─æß╗º payload |
| **NestJS (`nest-ai-service`)** | Lß║»ng nghe event tß╗½ RabbitMQ queue `ai.events` ΓåÆ Gß╗¡i email qua SendGrid ΓÇö **─É├ú ho├án th├ánh** Γ£à |

> **Hß╗ç thß╗æng sß╗¡ dß╗Ñng chung queue `ai.events`** (queue d├╣ng chung cho to├án bß╗Ö hß╗ç thß╗æng notification), kh├┤ng tß║ío th├¬m queue ri├¬ng ─æß╗â tr├ính ph├ón mß║únh cß║Ñu tr├║c RabbitMQ.

---

## 2. Luß╗ông hoß║ít ─æß╗Öng (Event Flow)

```
[PostgreSQL] ΓåÉ Go Core Service cß║¡p nhß║¡t bß║úng ownership
     Γöé
     Γöé  Publish RabbitMQ event:
     Γöé  Exchange: "product-trace.events"
     Γöé  Routing key: "ownership.transferred"
     Γöé  Payload: { event_id, event_type, payload: { email, full_name, product_name } }
     Γû╝
RabbitMQ Queue: "ai.events"
     Γöé
     Γû╝
NotificationConsumer (NestJS Worker)  ΓåÉ lß║»ng nghe li├¬n tß╗Ñc
     Γöé  case "ownership.transferred"
     Γû╝
MailService.sendOwnershipTransferredEmail()
     Γöé  Gß╗ìi SendGrid API vß╗¢i templateId + dynamicTemplateData
     Γû╝
Email ─æß║┐n h├▓m th╞░ kh├ích h├áng  Γ£à
```

---

## 3. C├íc file ─æ├ú thay ─æß╗òi trong `nest-ai-service`

| File | M├┤ tß║ú thay ─æß╗òi |
|---|---|
| `src/messaging/rabbitmq/rabbitmq.constants.ts` | Th├¬m `OWNERSHIP_TRANSFERRED: 'ownership.transferred'` v├áo `ROUTING_KEYS` v├á `EVENT_TYPES` |
| `src/messaging/rabbitmq/rabbitmq.service.ts` | Th├¬m `OWNERSHIP_TRANSFERRED` v├áo mß║úng `routingKeys` ─æß╗â tß╗▒ ─æß╗Öng bind routing key mß╗¢i v├áo queue `ai.events` khi khß╗ƒi tß║ío |
| `src/messaging/consumers/notification.consumer.ts` | Th├¬m `case RABBITMQ.EVENT_TYPES.OWNERSHIP_TRANSFERRED` trong switch ─æß╗â gß╗ìi `sendOwnershipTransferredEmail` |
| `src/modules/mail/mail.service.ts` | Th├¬m method `sendOwnershipTransferredEmail(to, fullName, productName)` ΓÇö gß╗¡i qua SendGrid Dynamic Template hoß║╖c fallback HTML thuß║ºn |
| `.env.example` | Th├¬m `OWNERSHIP_TRANSFERRED_TEMPLATE_ID` c├╣ng to├án bß╗Ö template IDs thß╗▒c tß║┐ ─æang d├╣ng |
| `docker-compose.yml` | Th├¬m biß║┐n `OWNERSHIP_TRANSFERRED_TEMPLATE_ID` v├áo environment cß╗ºa `nest-ai-service` v├á sß╗¡a `RABBITMQ_URL` ─æ├║ng ─æß╗ïa chß╗ë container |

---

## 4. Cß║Ñu tr├║c Payload RabbitMQ

Go Core Service **bß║»t buß╗Öc** publish ─æ├║ng schema sau:

```json
{
  "event_id": "uuid-duy-nhat",
  "event_type": "ownership.transferred",
  "payload": {
    "email": "nguoidung@gmail.com",
    "full_name": "Nguyß╗àn V─ân A",
    "product_name": "iPhone 16 Pro Max 256GB"
  }
}
```

> **L╞░u ├╜**: Nß║┐u thiß║┐u `full_name` hoß║╖c `product_name`, hß╗ç thß╗æng c├│ gi├í trß╗ï fallback mß║╖c ─æß╗ïnh (`'User'`, `'Sß║ún phß║⌐m cß╗ºa bß║ín'`) ─æß╗â tr├ính crash ΓÇö nh╞░ng email sß║╜ thiß║┐u th├┤ng tin.

---

## 5. Cß║Ñu h├¼nh SendGrid Template

**Template ID**: `d-1f78adcf1e3644e6bee66c2ea402af69`

| T├¬n biß║┐n SendGrid | Dß╗» liß╗çu truyß╗ün v├áo | V├¡ dß╗Ñ |
|---|---|---|
| `{{fullName}}` | T├¬n ng╞░ß╗¥i nhß║¡n sß╗ƒ hß╗»u | `Nguyß╗àn V─ân A` |
| `{{productName}}` | T├¬n sß║ún phß║⌐m ─æ╞░ß╗úc chuyß╗ân | `iPhone 16 Pro Max 256GB` |
| `{{frontendUrl}}` | Link hß╗ç thß╗æng | `http://localhost:5173` |
| `{{year}}` | N─âm hiß╗çn tß║íi (auto) | `2026` |

---

## 6. Cß║Ñu h├¼nh m├┤i tr╞░ß╗¥ng (.env)

Th├¬m biß║┐n sau v├áo file `.env` (─æ├ú c├│ sß║╡n trong `.env.example`):

```env
OWNERSHIP_TRANSFERRED_TEMPLATE_ID=d-1f78adcf1e3644e6bee66c2ea402af69
```

> Nß║┐u `OWNERSHIP_TRANSFERRED_TEMPLATE_ID` kh├┤ng ─æ╞░ß╗úc set, hß╗ç thß╗æng d├╣ng Template ID tr├¬n l├ám mß║╖c ─æß╗ïnh (─æ├ú hardcode fallback trong code).
> Nß║┐u `SENDGRID_API_KEY` kh├┤ng ─æ╞░ß╗úc set, hß╗ç thß╗æng chß║íy ß╗ƒ **MOCK mode** ΓÇö chß╗ë log ra console.

---

## 7. H╞░ß╗¢ng dß║½n Test (Tester / Dev kh├íc)

### Test thß╗º c├┤ng qua RabbitMQ Management UI
1. Mß╗ƒ `http://localhost:15672` (─æ─âng nhß║¡p: `guest` / `guest`).
2. V├áo **Exchanges** ΓåÆ chß╗ìn exchange `product-trace.events`.
3. T├¼m mß╗Ñc **Publish message**, ─æiß╗ün:
   - **Routing key**: `ownership.transferred`
   - **Payload**:
     ```json
     {
       "event_id": "test-ownership-001",
       "event_type": "ownership.transferred",
       "payload": {
         "email": "test@gmail.com",
         "full_name": "Nguyß╗àn V─ân A",
         "product_name": "iPhone 16 Pro Max 256GB"
       }
     }
     ```
4. Nhß║Ñn **Publish message**.
5. Kiß╗âm tra log cß╗ºa `nest-ai-service`:
   ```
   [NotificationConsumer] Ownership transferred email sent to test@gmail.com
   [NotificationConsumer] [Success] Event Type: ownership.transferred | Acknowledged.
   ```
6. Kiß╗âm tra h├▓m th╞░ nhß║¡n email vß╗¢i giao diß╗çn tß╗½ SendGrid template.

---

## 8. ─Éiß╗âm quan trß╗ìng cho Dev tiß║┐p nhß║¡n

### Dev Go Core Service cß║ºn l├ám:
- Sau khi chuyß╗ân quyß╗ün sß╗ƒ hß╗»u th├ánh c├┤ng trong PostgreSQL, publish event l├¬n RabbitMQ vß╗¢i:
  - **Exchange**: `product-trace.events`
  - **Routing key**: `ownership.transferred`
  - **Payload**: JSON ─æ├║ng schema mß╗Ñc 4 ß╗ƒ tr├¬n
- Th├¬m constant `OwnershipTransferredRK = "ownership.transferred"` v├áo `apps/go-core-service/internal/events/rabbitmq/constants.go`

### Dev NestJS KH├öNG cß║ºn sß╗¡a g├¼ th├¬m khi:
- Go Core Service thay ─æß╗òi nß╗Öi dung payload ΓÇö chß╗ë cß║ºn ─æß║úm bß║úo `email`, `full_name`, `product_name` trong `payload` object.

### ─Éß╗â th├¬m loß║íi th├┤ng b├ío mß╗¢i:
1. Th├¬m constant mß╗¢i v├áo `rabbitmq.constants.ts`
2. Bind routing key mß╗¢i v├áo `rabbitmq.service.ts`
3. Th├¬m `case` mß╗¢i v├áo `notification.consumer.ts`
4. Th├¬m method mß╗¢i v├áo `mail.service.ts`

---

# T├ái liß╗çu B├án giao: Module Th├┤ng B├ío Hß║┐t Hß║ín Bß║úo H├ánh (Warranty Expired Worker)

## Tß╗òng quan

─É├ú triß╗ân khai ho├án chß╗ënh t├¡nh n─âng gß╗¡i **email th├┤ng b├ío hß║┐t hß║ín bß║úo h├ánh** cho kh├ích h├áng th├┤ng qua hß╗ç thß╗æng Event-Driven (RabbitMQ ΓåÆ NestJS Worker ΓåÆ SendGrid).

- **Use Case**: Gß╗¡i email th├┤ng b├ío cho kh├ích h├áng khi sß║ún phß║⌐m cß╗ºa hß╗ì ─æ├ú hß║┐t hß║ín bß║úo h├ánh.
- **Routing Key / Event Type**: `warranty.expired`
- **Queue**: `ai.events` (Sß╗¡ dß╗Ñng chung queue vß╗¢i c├íc th├┤ng b├ío kh├íc)
- **SendGrid Template ID**: `d-ded28a6c91104c11bc548b08002c74f5`
- **Trß║íng th├íi**: Γ£à **─É├ú test th├ánh c├┤ng end-to-end** ΓÇö Email gß╗¡i tß╗¢i `hoangnguyen280004@gmail.com`

---

## 1. Cß║Ñu tr├║c Payload RabbitMQ

Go Core Service hoß║╖c Scheduler Worker sß║╜ publish event l├¬n RabbitMQ vß╗¢i routing key `warranty.expired` v├á schema payload sau:

```json
{
  "event_id": "uuid-duy-nhat",
  "event_type": "warranty.expired",
  "payload": {
    "email": "nguoidung@gmail.com",
    "full_name": "Nguyß╗àn V─ân A",
    "product_name": "iPhone 15 Pro Max 256GB",
    "warranty_end_date": "15/07/2026"
  }
}
```

---

## 2. C├íc file ─æ├ú thay ─æß╗òi trong `nest-ai-service`

| File | M├┤ tß║ú thay ─æß╗òi |
|---|---|
| `src/messaging/rabbitmq/rabbitmq.constants.ts` | Th├¬m `WARRANTY_EXPIRED: 'warranty.expired'` v├áo `ROUTING_KEYS` v├á `EVENT_TYPES` |
| `src/messaging/rabbitmq/rabbitmq.service.ts` | Th├¬m `WARRANTY_EXPIRED` v├áo danh s├ích `routingKeys` ─æß╗â tß╗▒ ─æß╗Öng bind v├áo queue khi khß╗ƒi chß║íy |
| `src/messaging/consumers/notification.consumer.ts` | Th├¬m case `WARRANTY_EXPIRED` ─æß╗â gß╗ìi method `sendWarrantyExpiredEmail` |
| `src/modules/mail/mail.service.ts` | Th├¬m method `sendWarrantyExpiredEmail` hß╗ù trß╗ú cß║ú SendGrid template v├á HTML fallback |
| `.env.example` & `.env` | Th├¬m cß║Ñu h├¼nh biß║┐n m├┤i tr╞░ß╗¥ng `WARRANTY_EXPIRED_TEMPLATE_ID=d-ded28a6c91104c11bc548b08002c74f5` |
| `docker-compose.yml` | Truyß╗ün biß║┐n `WARRANTY_EXPIRED_TEMPLATE_ID` v├áo m├┤i tr╞░ß╗¥ng cß╗ºa container `nest-ai-service` |

---

## 3. Cß║Ñu h├¼nh SendGrid Template

**Template ID**: `d-ded28a6c91104c11bc548b08002c74f5`

| T├¬n biß║┐n SendGrid | Dß╗» liß╗çu truyß╗ün v├áo | V├¡ dß╗Ñ |
|---|---|---|
| `{{fullName}}` | T├¬n kh├ích h├áng | `Nguyß╗àn V─ân A` |
| `{{productName}}` | T├¬n sß║ún phß║⌐m | `iPhone 15 Pro Max 256GB` |
| `{{endDate}}` | Ng├áy hß║┐t hß║ín bß║úo h├ánh | `15/07/2026` |
| `{{frontendUrl}}` | Link hß╗ç thß╗æng | `http://localhost:5173/warranty` |
| `{{year}}` | N─âm hiß╗çn tß║íi (auto) | `2026` |

---

## 4. Cß║Ñu h├¼nh m├┤i tr╞░ß╗¥ng (.env)

```env
WARRANTY_EXPIRED_TEMPLATE_ID=d-ded28a6c91104c11bc548b08002c74f5
```

---

## 5. Test thß╗º c├┤ng qua RabbitMQ Management UI

1. Mß╗ƒ `http://localhost:15672` (─æ─âng nhß║¡p: `guest` / `guest` hoß║╖c `admin` / `admin123`).
2. V├áo **Exchanges** ΓåÆ chß╗ìn exchange `product-trace.events`.
3. T├¼m mß╗Ñc **Publish message**, ─æiß╗ün:
   - **Routing key**: `warranty.expired`
   - **Payload**:
     ```json
     {
       "event_id": "test-warranty-expired-001",
       "event_type": "warranty.expired",
       "payload": {
         "email": "test@gmail.com",
         "full_name": "Nguyß╗àn V─ân A",
         "product_name": "iPhone 15 Pro Max 256GB",
         "warranty_end_date": "15/07/2026"
       }
     }
     ```
4. Nhß║Ñn **Publish message**.
5. Kiß╗âm tra log cß╗ºa `nest-ai-service`:
   ```
   [NotificationConsumer] Warranty expired email sent to test@gmail.com
   [NotificationConsumer] [Success] Event Type: warranty.expired | Acknowledged.
   ```
6. Kiß╗âm tra hß╗Öp th╞░ (bao gß╗ôm cß║ú **Th╞░ r├íc / Spam**).

---

## 6. ΓÜá∩╕Å L╞░u ├╜ quan trß╗ìng: SendGrid Sender Verification

> **Email c├│ thß╗â v├áo th╞░ r├íc nß║┐u sender ch╞░a ─æ╞░ß╗úc x├íc minh!**

─Éß╗â email gß╗¡i v├áo **Hß╗Öp th╞░ ─æß║┐n** (kh├┤ng phß║úi spam), cß║ºn x├íc minh sender trong SendGrid:

1. V├áo **[https://app.sendgrid.com/settings/sender_auth](https://app.sendgrid.com/settings/sender_auth)**
2. Chß╗ìn **"Verify a Single Sender"**
3. ─Éiß╗ün ─æß╗ïa chß╗ë email ng╞░ß╗¥i gß╗¡i (v├¡ dß╗Ñ: `nguyenhoang280004@gmail.com`)
4. Nhß║Ñn **Create** ΓåÆ SendGrid gß╗¡i email x├íc nhß║¡n
5. Mß╗ƒ email x├íc nhß║¡n v├á click **"Verify Single Sender"**

Sau khi x├íc minh, email sß║╜ gß╗¡i v├áo hß╗Öp th╞░ ─æß║┐n thay v├¼ spam.

---

## 7. Lß╗ïch sß╗¡ thay ─æß╗òi

| Ng├áy | Nß╗Öi dung |
|---|---|
| 15/07/2026 | Khß╗ƒi tß║ío module, implement `sendOwnershipTransferredEmail` vß╗¢i fallback HTML thuß║ºn |
| 15/07/2026 | Th├¬m constant `OWNERSHIP_TRANSFERRED` v├áo `rabbitmq.constants.ts` v├á bind queue |
| 15/07/2026 | Th├¬m `case` xß╗¡ l├╜ event `ownership.transferred` v├áo `notification.consumer.ts` |
| 15/07/2026 | Th├¬m `OWNERSHIP_TRANSFERRED_TEMPLATE_ID` v├áo `.env.example` v├á `docker-compose.yml` |
| 15/07/2026 | Test th├ánh c├┤ng end-to-end qua Docker ΓÇö email gß╗¡i tß╗¢i `nguyenzc17@gmail.com` Γ£à |
| 15/07/2026 | Push code l├¬n nh├ính `feature/notification-ownership`, tß║ío PR ΓåÆ `develop` |
| 15/07/2026 | Triß╗ân khai **Warranty Expired Worker** ΓÇö routing key `warranty.expired`, method `sendWarrantyExpiredEmail` |
| 15/07/2026 | Cß║Ñu h├¼nh `WARRANTY_EXPIRED_TEMPLATE_ID=d-ded28a6c91104c11bc548b08002c74f5` |
| 15/07/2026 | Test end-to-end th├ánh c├┤ng: RabbitMQ ΓåÆ NestJS ΓåÆ SendGrid ΓåÆ `hoangnguyen280004@gmail.com` Γ£à |
| 15/07/2026 | Ghi nhß║¡n: email v├áo spam do sender ch╞░a verify ΓÇö h╞░ß╗¢ng dß║½n Sender Verification tß║íi mß╗Ñc 6 |



