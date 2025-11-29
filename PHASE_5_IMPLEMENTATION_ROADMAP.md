# PHASE 5: DETAILED IMPLEMENTATION ROADMAP

**Platform**: VTP (Video Teaching Platform)  
**Target Users**: Syrian Students (Arabic-first)  
**Duration**: 4-6 weeks (full sprint)  
**Quality Target**: Production-grade, matching backend sophistication  

---

## Week 1: Foundation & Authentication (Days 1-10)

### Days 1-3: Project Initialization

**Day 1: Project Setup**
```bash
# Initialize Next.js project with TypeScript
npx create-next-app@latest vtp-frontend --typescript --tailwind

# Install core dependencies
npm install react-hook-form zod @hookform/resolvers
npm install zustand axios
npm install next-i18next i18next react-i18next
npm install @radix-ui/react-primitive
npm install react-toastify recharts

# Install shadcn/ui CLI
npm install -g shadcn-ui
```

**Project Structure Setup**:
```
vtp-frontend/
├── src/
│   ├── components/
│   ├── pages/
│   ├── hooks/
│   ├── services/
│   ├── store/
│   ├── types/
│   ├── utils/
│   ├── styles/
│   └── i18n/
├── public/
├── tests/
└── docker-compose.yml
```

**Files to Create**:
1. `src/utils/constants.ts` - API URLs, app constants
2. `src/types/index.ts` - TypeScript definitions
3. `src/services/api.ts` - Axios setup
4. `.env.local` - Environment variables
5. `tsconfig.json` - TypeScript config
6. `next.config.js` - Next.js config

**Deliverables**: ✅ Project structure, dependencies installed, ready for components

---

**Day 2: Design System & Base Components**

**Install Shadcn/ui Components**:
```bash
# Button, Card, Input, Form, Select, Dialog, Modal
npx shadcn-ui@latest add button
npx shadcn-ui@latest add card
npx shadcn-ui@latest add input
npx shadcn-ui@latest add form
npx shadcn-ui@latest add select
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add modal
npx shadcn-ui@latest add badge
npx shadcn-ui@latest add alert
npx shadcn-ui@latest add separator
npx shadcn-ui@latest add avatar
npx shadcn-ui@latest add dropdown-menu
npx shadcn-ui@latest add tabs
# (Add 20+ total components)
```

**Files to Create**:
1. `src/components/common/Button.tsx` - Custom Button wrapper
2. `src/components/common/Card.tsx` - Custom Card wrapper
3. `src/components/common/Input.tsx` - Custom Input wrapper
4. `src/components/common/Modal.tsx` - Modal component
5. `src/components/common/Loader.tsx` - Loading spinner
6. `src/components/common/Error.tsx` - Error display
7. `src/components/common/Success.tsx` - Success message

**Tests to Write** (10):
- ButtonComponent.test.tsx
- CardComponent.test.tsx
- InputComponent.test.tsx
- ModalComponent.test.tsx
- Loader.test.tsx
- Error.test.tsx
- Success.test.tsx
- FormValidation.test.tsx
- SelectComponent.test.tsx
- BadgeComponent.test.tsx

**Deliverables**: ✅ Design system complete, 20+ shadcn/ui components imported, base components wrapped

---

**Day 3: Arabic Localization Setup**

**Files to Create**:
1. `src/i18n/en.json` - English translations (200+ keys)
2. `src/i18n/ar.json` - Arabic translations (200+ keys)
3. `src/i18n/config.ts` - i18next configuration
4. `src/components/common/LanguageSwitcher.tsx` - Language toggle
5. `src/styles/rtl.css` - RTL-specific styles
6. `next.config.js` - Add i18n configuration

**Translation Keys** (Sample):
```json
{
  "nav": {
    "home": "الرئيسية",
    "courses": "الدورات",
    "dashboard": "لوحة التحكم",
    "profile": "الملف الشخصي"
  },
  "auth": {
    "login": "تسجيل الدخول",
    "register": "إنشاء حساب",
    "email": "البريد الإلكتروني",
    "password": "كلمة المرور"
  },
  "errors": {
    "required": "هذا الحقل مطلوب",
    "invalid_email": "البريد الإلكتروني غير صحيح",
    "password_mismatch": "كلمات المرور غير متطابقة"
  }
}
```

**RTL Styling** (Tailwind Config):
```javascript
// tailwind.config.js
module.exports = {
  plugins: [require('tailwindcss-rtl')],
  theme: {
    extend: {},
  }
}
```

**Deliverables**: ✅ i18n setup complete, 200+ Arabic translations, RTL styling configured

---

### Days 4-5: Authentication UI

**Files to Create** (Auth Components):

1. **`src/components/auth/LoginForm.tsx`** (150 lines)
   - Email input field
   - Password input field
   - "Remember me" checkbox
   - "Forgot password?" link
   - Submit button with loading state
   - Error message display
   - Form validation with Zod

2. **`src/components/auth/RegisterForm.tsx`** (180 lines)
   - Full name input
   - Email input
   - Password input
   - Confirm password input
   - Terms agreement checkbox
   - Submit button
   - Link to login page
   - Form validation

3. **`src/components/auth/PasswordResetForm.tsx`** (150 lines)
   - Email input (for forgot password)
   - Submit button
   - Success message after submission

4. **`src/components/auth/ResetPasswordForm.tsx`** (150 lines)
   - New password input
   - Confirm password input
   - Submit button
   - Validation rules (min 8 chars, uppercase, number)

5. **`src/pages/auth/login.tsx`** (100 lines)
   - LoginForm component
   - "New user? Register here" link
   - OAuth buttons (future)

6. **`src/pages/auth/register.tsx`** (100 lines)
   - RegisterForm component
   - "Already have account? Login" link

7. **`src/pages/auth/forgot-password.tsx`** (80 lines)
   - PasswordResetForm component
   - Email sent confirmation

8. **`src/pages/auth/reset-password/[token].tsx`** (100 lines)
   - ResetPasswordForm component
   - Token validation

**Auth Services** (API Integration):

1. **`src/services/auth.ts`** (200 lines)
   ```typescript
   export const authService = {
     login: async (email: string, password: string) => { /* ... */ },
     register: async (data: RegisterData) => { /* ... */ },
     logout: () => { /* ... */ },
     refreshToken: async () => { /* ... */ },
     verifyEmail: async (token: string) => { /* ... */ },
     forgotPassword: async (email: string) => { /* ... */ },
     resetPassword: async (token: string, password: string) => { /* ... */ },
     getProfile: async () => { /* ... */ },
     updateProfile: async (data: ProfileData) => { /* ... */ },
     changePassword: async (oldPassword: string, newPassword: string) => { /* ... */ }
   }
   ```

2. **`src/store/authStore.ts`** (150 lines) - Zustand store
   ```typescript
   export const useAuthStore = create((set) => ({
     user: null,
     token: null,
     isAuthenticated: false,
     login: async (email, password) => { /* ... */ },
     logout: () => { /* ... */ },
     setUser: (user) => set({ user }),
     setToken: (token) => set({ token })
   }))
   ```

3. **`src/hooks/useAuth.ts`** (100 lines)
   ```typescript
   export const useAuth = () => {
     const { user, token, login, logout } = useAuthStore()
     const { push } = useRouter()
     return { user, token, login, logout, isLoading, error }
   }
   ```

**Protected Routes**:

1. **`src/components/common/ProtectedRoute.tsx`** (80 lines)
   ```typescript
   export const ProtectedRoute = ({ children }) => {
     const { isAuthenticated } = useAuth()
     const router = useRouter()
     
     useEffect(() => {
       if (!isAuthenticated) router.push('/auth/login')
     }, [isAuthenticated])
     
     return isAuthenticated ? children : <Loader />
   }
   ```

**Tests to Write** (20):
- LoginForm.test.tsx (5 tests: render, validation, submit, error)
- RegisterForm.test.tsx (5 tests)
- PasswordResetForm.test.tsx (3 tests)
- authService.test.ts (4 tests: API calls)
- useAuth.test.ts (3 tests: hook behavior)

**API Endpoints Used**:
```
POST   /auth/register
POST   /auth/login
POST   /auth/verify-email
POST   /auth/forgot-password
POST   /auth/reset-password
POST   /auth/refresh-token
GET    /auth/profile
PUT    /auth/profile
POST   /auth/change-password
```

**Deliverables**: ✅ Auth UI complete, forms validating, API integration working, protected routes functional

---

## Week 2: Core Features (Days 11-20)

### Days 6-8: Live Streaming Interface

**Files to Create** (WebRTC Integration):

1. **`src/services/streaming.ts`** (300+ lines)
   ```typescript
   export const streamingService = {
     createRoom: async (lectureId: UUID) => { /* POST /signaling/room/create */ },
     joinRoom: async (roomId: string) => { /* POST /signaling/room/join */ },
     leaveRoom: async (roomId: string) => { /* POST /signaling/room/leave */ },
     createProducer: async (roomId: string, kind: 'audio'|'video') => { /* ... */ },
     createConsumer: async (roomId: string, producerId: string) => { /* ... */ },
     pauseProducer: async (producerId: string) => { /* ... */ },
     resumeProducer: async (producerId: string) => { /* ... */ },
     getRTPParameters: async () => { /* GET /signaling/rtpParameters */ }
   }
   ```

2. **`src/hooks/useWebRTC.ts`** (250 lines)
   ```typescript
   export const useWebRTC = (roomId: string) => {
     const [localStream, setLocalStream] = useState(null)
     const [remoteStreams, setRemoteStreams] = useState([])
     const [isAudioEnabled, setIsAudioEnabled] = useState(true)
     const [isVideoEnabled, setIsVideoEnabled] = useState(true)
     
     // Initialize WebRTC, create producer, add consumers
     return {
       localStream,
       remoteStreams,
       toggleAudio: () => { /* ... */ },
       toggleVideo: () => { /* ... */ },
       shareScreen: () => { /* ... */ },
       leaveRoom: () => { /* ... */ }
     }
   }
   ```

3. **`src/components/streaming/VideoGrid.tsx`** (200 lines)
   - Responsive grid layout (1, 2, 4, 9 videos)
   - Video element per participant
   - Instructor video highlighted
   - Mobile responsive (stacked layout)

4. **`src/components/streaming/RemoteVideo.tsx`** (100 lines)
   - Display remote participant video
   - Participant name/badge
   - Connection status indicator
   - Audio level indicator

5. **`src/components/streaming/LocalVideo.tsx`** (80 lines)
   - Preview of own video
   - "You" badge
   - Mic/camera disabled indicators

6. **`src/components/streaming/ControlPanel.tsx`** (150 lines)
   - Mic toggle button
   - Camera toggle button
   - Screen share button
   - Leave room button
   - Settings menu

7. **`src/components/streaming/ParticipantList.tsx`** (120 lines)
   - List of participants
   - Online status indicator
   - Muted indicator
   - Hand raised indicator (for Q&A)

8. **`src/components/streaming/ChatPanel.tsx`** (150 lines)
   - Message list
   - Message input
   - Timestamp for messages
   - User avatars
   - RTL message alignment

9. **`src/pages/streaming/room/[id].tsx`** (150 lines)
   - Layout: VideoGrid on left, ChatPanel on right
   - Responsive: mobile stacks vertically
   - Error handling and reconnection

10. **`src/pages/streaming/setup.tsx`** (100 lines)
    - Camera/microphone permission checks
    - Device selector (multiple cameras/mics)
    - Bandwidth test
    - Network quality indicator
    - "Join room" button

**Store for Streaming State** (Zustand):

1. **`src/store/streamingStore.ts`** (150 lines)
   ```typescript
   export const useStreamingStore = create((set) => ({
     roomId: null,
     participants: [],
     isAudioEnabled: true,
     isVideoEnabled: true,
     networkQuality: 'good',
     setRoomId: (id) => set({ roomId: id }),
     addParticipant: (p) => set((s) => ({ participants: [...s.participants, p] })),
     removeParticipant: (id) => set((s) => ({ 
       participants: s.participants.filter(p => p.id !== id) 
     }))
   }))
   ```

**Tests to Write** (15):
- VideoGrid.test.tsx (layout responsive tests)
- RemoteVideo.test.tsx
- LocalVideo.test.tsx
- ControlPanel.test.tsx (button interactions)
- ParticipantList.test.tsx
- ChatPanel.test.tsx
- useWebRTC.test.ts (WebRTC hook logic)
- streamingService.test.ts (API calls)

**Deliverables**: ✅ Live streaming UI complete, WebRTC integration working, video grid responsive

---

### Days 9-10: Video Playback & Player

**Files to Create** (HLS Video Player):

1. **`src/components/video/VideoPlayer.tsx`** (300+ lines)
   ```typescript
   export const VideoPlayer = ({ src, onTimeUpdate, onEnded }) => {
     const videoRef = useRef(null)
     const [isPlaying, setIsPlaying] = useState(false)
     const [currentTime, setCurrentTime] = useState(0)
     const [duration, setDuration] = useState(0)
     const [quality, setQuality] = useState('auto')
     const [playbackRate, setPlaybackRate] = useState(1)
     const [isFullscreen, setIsFullscreen] = useState(false)
     
     // Initialize HLS.js for adaptive streaming
     // Handle quality switching, playback speed, fullscreen
     // Track watch time for analytics
   }
   ```

2. **`src/hooks/useVideoPlayer.ts`** (200 lines)
   ```typescript
   export const useVideoPlayer = (videoElement, hlsUrl) => {
     // Initialize HLS.js instance
     // Manage playback state
     // Track watch time
     // Handle quality levels
     return {
       play: () => { /* ... */ },
       pause: () => { /* ... */ },
       seek: (time) => { /* ... */ },
       setQuality: (level) => { /* ... */ },
       setPlaybackRate: (rate) => { /* ... */ },
       toggleFullscreen: () => { /* ... */ }
     }
   }
   ```

3. **`src/components/video/PlaybackControls.tsx`** (150 lines)
   - Play/pause button
   - Progress bar with seek
   - Duration/current time display
   - Volume control
   - Fullscreen button
   - Settings button

4. **`src/components/video/QualitySelector.tsx`** (100 lines)
   - Dropdown with quality options
   - Auto quality toggle
   - Current quality indicator
   - Bandwidth indicator

5. **`src/components/video/PlaybackSpeedSelector.tsx`** (80 lines)
   - Speed options (0.5x, 0.75x, 1x, 1.25x, 1.5x, 2x)
   - Currently selected indicator

6. **`src/components/video/RecordingList.tsx`** (180 lines)
   - Grid of recording cards
   - Search input
   - Filter by course/date
   - Sort options

7. **`src/components/video/RecordingCard.tsx`** (120 lines)
   - Thumbnail image
   - Title
   - Duration
   - Date
   - Course name
   - Completion status
   - Play button

8. **`src/pages/playback/[id].tsx`** (100 lines)
   - VideoPlayer component
   - Recording metadata (title, instructor, date)
   - Course info sidebar
   - Next/previous lecture buttons

9. **`src/pages/recordings.tsx`** (100 lines)
   - RecordingList component
   - Search and filters
   - Sorting

10. **`src/services/playback.ts`** (150 lines)
    ```typescript
    export const playbackService = {
      getRecordings: async (courseId?: UUID) => { /* GET /recordings */ },
      getRecording: async (id: UUID) => { /* GET /recordings/{id} */ },
      getPlaybackStream: async (id: UUID) => { /* GET /playback/{id}/stream */ },
      startPlayback: async (id: UUID) => { /* POST /playback/{id}/start */ },
      trackWatchTime: async (recordingId: UUID, position: number) => { /* ... */ }
    }
    ```

**Watch Time Tracking**:
1. **`src/hooks/useWatchTime.ts`** (100 lines)
   - Track current playback position
   - Send analytics events to backend
   - Mark as watched at 80%+ completion
   - Store position in localStorage for resume

**Tests to Write** (15):
- VideoPlayer.test.tsx (5 tests: play, pause, seek, quality)
- PlaybackControls.test.tsx (3 tests)
- QualitySelector.test.tsx
- RecordingList.test.tsx
- RecordingCard.test.tsx
- useVideoPlayer.test.ts (3 tests)

**API Endpoints Used**:
```
GET    /playback/{id}/stream
POST   /playback/{id}/start
GET    /recordings
GET    /recordings/{id}
GET    /recordings/{id}/metadata
```

**Deliverables**: ✅ Video player functional, HLS streaming working, quality selection working, watch time tracking

---

## Week 3: Dashboard & Localization (Days 21-30)

### Days 11-13: Course Management

**Files to Create**:

1. **`src/components/courses/CourseList.tsx`** (150 lines)
   - Grid/list view toggle
   - Search functionality
   - Filter by status (enrolled, archived, all)
   - Sort options
   - Pagination

2. **`src/components/courses/CourseCard.tsx`** (120 lines)
   - Course cover image
   - Title
   - Instructor name
   - Student count
   - Progress bar
   - Enrollment status
   - "Enroll" or "View" button

3. **`src/components/courses/CourseDetails.tsx`** (200 lines)
   - Course header (cover, title, instructor, stats)
   - Course description
   - Lecture list
   - Enrollment button
   - Share button

4. **`src/components/courses/LectureList.tsx`** (150 lines)
   - Numbered list of lectures
   - Lecture title, duration
   - Completion indicator
   - Play button
   - Download button

5. **`src/components/courses/EnrollmentButton.tsx`** (100 lines)
   - "Enroll" button (if not enrolled)
   - Confirmation modal
   - Success notification
   - Loading state

6. **`src/pages/courses/index.tsx`** (100 lines)
   - CourseList component
   - Welcome message
   - "Create course" button (if instructor)

7. **`src/pages/courses/[id].tsx`** (150 lines)
   - CourseDetails component
   - LectureList component

8. **`src/pages/dashboard.tsx`** (150 lines)
   - "My Courses" section
   - Quick stats (enrolled, completion rate)
   - Recent activity
   - Recommended courses

9. **`src/pages/instructor/courses.tsx`** (150 lines) - Instructor view
   - "My Courses" list
   - "Create Course" button
   - Edit/delete options
   - Course statistics

10. **`src/pages/instructor/courses/create.tsx`** (150 lines)
    - Course form (title, description, cover upload)
    - Lecture upload
    - Publish course

11. **`src/services/courses.ts`** (200 lines)
    ```typescript
    export const courseService = {
      getCourses: async (filters?) => { /* GET /courses */ },
      getCourse: async (id: UUID) => { /* GET /courses/{id} */ },
      createCourse: async (data) => { /* POST /courses */ },
      updateCourse: async (id, data) => { /* PUT /courses/{id} */ },
      deleteCourse: async (id) => { /* DELETE /courses/{id} */ },
      enroll: async (courseId) => { /* POST /courses/{id}/enroll */ },
      getLectures: async (courseId) => { /* GET /courses/{id}/lectures */ }
    }
    ```

12. **`src/store/courseStore.ts`** (150 lines) - Zustand
    ```typescript
    export const useCourseStore = create((set) => ({
      courses: [],
      currentCourse: null,
      enrolledCourses: [],
      setCourses: (courses) => set({ courses }),
      setCurrentCourse: (course) => set({ currentCourse: course }),
      enrollCourse: (courseId) => { /* ... */ }
    }))
    ```

**Tests to Write** (15):
- CourseList.test.tsx
- CourseCard.test.tsx
- CourseDetails.test.tsx
- LectureList.test.tsx
- EnrollmentButton.test.tsx
- courseService.test.ts
- useCourseStore.test.ts

**API Endpoints Used**:
```
POST   /courses
GET    /courses
GET    /courses/{id}
PUT    /courses/{id}
POST   /courses/{id}/enroll
GET    /courses/{id}/lectures
```

**Deliverables**: ✅ Course management complete, enrollment working, student/instructor views functional

---

### Days 14-15: Analytics Dashboard

**Files to Create**:

1. **`src/components/analytics/Dashboard.tsx`** (200 lines)
   - Metrics overview (4 cards)
   - Engagement chart
   - Course progress chart
   - Recent alerts
   - Student list (instructor view)

2. **`src/components/analytics/MetricsCard.tsx`** (80 lines)
   - Title, value, unit
   - Trend indicator (↑↓)
   - Color coding (good/warning/danger)

3. **`src/components/analytics/EngagementChart.tsx`** (120 lines)
   - Line chart using Recharts
   - Time period selector (week/month)
   - Tooltip with data points

4. **`src/components/analytics/AlertList.tsx`** (100 lines)
   - List of alerts
   - Severity indicator (warning/critical)
   - Alert message
   - Timestamp
   - Dismiss button

5. **`src/components/analytics/StudentMetricsTable.tsx`** (150 lines)
   - Table of students with metrics
   - Sortable columns (name, engagement, completion)
   - Filtering
   - Export to CSV button

6. **`src/pages/analytics/my-metrics.tsx`** (100 lines)
   - Dashboard for current user
   - Personal metrics

7. **`src/pages/analytics/course/[id].tsx`** (100 lines)
   - Course-specific analytics
   - Instructor view: student list with metrics

8. **`src/pages/analytics/alerts.tsx`** (100 lines)
   - Full alert list
   - Filter by severity, date
   - Clear all button

9. **`src/pages/analytics/reports.tsx`** (100 lines)
   - Download engagement report (PDF)
   - Download performance report (CSV)
   - Report generation status

10. **`src/services/analytics.ts`** (200 lines)
    ```typescript
    export const analyticsService = {
      getMetrics: async (userId?: UUID) => { /* GET /api/analytics/metrics */ },
      getLectureStats: async (lectureId) => { /* GET /api/analytics/lecture */ },
      getCourseStats: async (courseId) => { /* GET /api/analytics/course */ },
      getAlerts: async (filters?) => { /* GET /api/analytics/alerts */ },
      getEngagementReport: async (courseId) => { /* GET /api/analytics/reports/engagement */ },
      getPerformanceReport: async (courseId) => { /* GET /api/analytics/reports/performance */ }
    }
    ```

11. **`src/store/analyticsStore.ts`** (150 lines) - Zustand
    ```typescript
    export const useAnalyticsStore = create((set) => ({
      metrics: null,
      alerts: [],
      reports: {},
      setMetrics: (m) => set({ metrics: m }),
      addAlert: (alert) => set((s) => ({ alerts: [...s.alerts, alert] })),
      clearAlerts: () => set({ alerts: [] })
    }))
    ```

12. **Real-time Alert Updates** (WebSocket):
    ```typescript
    export const useAlertSubscription = () => {
      useEffect(() => {
        const ws = new WebSocket('ws://localhost:8080/alerts')
        ws.onmessage = (event) => {
          const alert = JSON.parse(event.data)
          useAnalyticsStore.setState((s) => ({
            alerts: [alert, ...s.alerts]
          }))
          toast.warning(alert.message)
        }
        return () => ws.close()
      }, [])
    }
    ```

**Tests to Write** (15):
- Dashboard.test.tsx
- MetricsCard.test.tsx
- EngagementChart.test.tsx
- AlertList.test.tsx
- StudentMetricsTable.test.tsx
- analyticsService.test.ts
- useAnalyticsStore.test.ts

**API Endpoints Used**:
```
GET    /api/analytics/metrics
GET    /api/analytics/lecture
GET    /api/analytics/course
GET    /api/analytics/alerts
GET    /api/analytics/reports/engagement
GET    /api/analytics/reports/performance
```

**Deliverables**: ✅ Analytics dashboard functional, charts rendering, alerts displaying, reports downloadable

---

### Days 16-17: Arabic Localization Completion

**Tasks**:
1. Complete 200+ translation keys
2. Professional review of Arabic translations
3. Set up right-to-left (RTL) CSS
4. Test language switching
5. Configure date/time localization

**Files**:
- `src/i18n/ar.json` - Update with all 200+ keys
- `src/i18n/en.json` - Complete English translations
- `src/styles/rtl.css` - RTL utilities
- `src/components/common/LanguageSwitcher.tsx` - Language toggle

**Arabic Translation Categories**:
- Navigation & menus (50 keys)
- Authentication (40 keys)
- Course management (50 keys)
- Streaming (40 keys)
- Analytics (30 keys)
- Error messages (50 keys)

**RTL Testing Checklist**:
- ✅ Form layouts (labels, inputs, buttons)
- ✅ Navigation (sidebar, header)
- ✅ Text direction (all text right-aligned)
- ✅ Margin/padding (left/right swapped)
- ✅ Icons (some need flipping)
- ✅ Mobile responsiveness (RTL on mobile)

**Deliverables**: ✅ Full Arabic localization complete, RTL styling perfect, all text translated

---

## Week 4: Testing & Deployment (Days 31-40)

### Days 18-19: Testing Infrastructure

**Unit Tests** (50+ tests):
- Components: 30 tests
- Hooks: 10 tests
- Utilities: 10 tests

**Integration Tests** (20+ tests):
- Auth flow (login → protected page → logout)
- Course flow (list → details → enroll)
- Video flow (list → play → track)
- Analytics flow (load metrics → view alerts)

**E2E Tests** (10+ scenarios):
- Complete user journey
- Instructor course creation
- Error recovery
- Mobile responsiveness

**Mock API Setup**:
- MSW handlers for all 53 endpoints
- Error scenarios (500, 404, timeout)
- Delay simulation

**Deliverables**: ✅ 100+ tests written, 80%+ code coverage, all flows tested

---

### Days 20-21: Deployment & DevOps

**Docker Setup**:
```dockerfile
# Multi-stage build
FROM node:18-alpine AS builder
WORKDIR /app
COPY package.json ./
RUN npm install
COPY . .
RUN npm run build

FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./
RUN npm install --production
EXPOSE 3000
CMD ["npm", "start"]
```

**Environment Configuration**:
- `.env.local` - Development
- `.env.staging` - Staging
- `.env.production` - Production

**CI/CD Pipeline** (GitHub Actions):
- Run tests on PR
- Build on merge
- Deploy to staging/production

**Performance Optimization**:
- Code splitting
- Image optimization
- Bundle analysis
- Lighthouse audit

**Deliverables**: ✅ Docker setup complete, CI/CD pipeline working, production deployment ready

---

## Final Deliverables Summary

### Week 1: Foundation
✅ Project initialized
✅ Design system (30+ shadcn/ui components)
✅ Arabic localization setup
✅ Auth UI (login, register, password reset)
✅ Protected routes

### Week 2: Core Features
✅ Live streaming (WebRTC, video grid, controls)
✅ Video playback (HLS player, quality, playback speed)
✅ Watch time tracking
✅ Recording list

### Week 3: Dashboard
✅ Course management (list, details, enrollment)
✅ Analytics dashboard (metrics, charts, alerts)
✅ Instructor view (course creation, student analytics)
✅ Full Arabic localization + RTL

### Week 4: Testing & Deployment
✅ 100+ unit/integration/E2E tests
✅ Docker containerization
✅ CI/CD pipeline
✅ Production deployment ready

---

## Success Metrics

### Code Quality
- ✅ TypeScript strict mode
- ✅ ESLint zero warnings
- ✅ Prettier formatting
- ✅ 80%+ test coverage

### Performance
- ✅ Page load < 2s
- ✅ Time to interactive < 3s
- ✅ Lighthouse score > 90
- ✅ Bundle size < 500KB (gzipped)

### User Experience
- ✅ All 53 endpoints integrated
- ✅ 100+ UI components
- ✅ Full Arabic support
- ✅ Mobile responsive

### Features
- ✅ Authentication complete
- ✅ Live streaming functional
- ✅ Video playback working
- ✅ Analytics dashboard live
- ✅ Course management operational

---

## Ready to Start?

**Recommended Approach**:
1. ✅ Review this plan
2. ✅ Approve scope & timeline
3. ✅ Start Phase 5A (Days 1-3)
4. ✅ Iterate through weeks

**What would you like to do?**
- Start Phase 5A immediately
- Adjust timeline/scope
- Modify technology choices
- Add/remove features
- Create day-by-day detailed tasks

Let me know and we'll begin Phase 5 frontend development! 🚀
