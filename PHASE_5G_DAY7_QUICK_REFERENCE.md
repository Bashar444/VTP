# Phase 5G Day 7 - Quick Reference Guide

## 📦 What Was Built

**4 React Components + 1 Service Layer**

### Components Location
```
vtp-frontend/src/components/
├── NetworkStatus.tsx & .css          (280 + 350 lines)
├── QualitySelector.tsx & .css        (310 + 350 lines)
├── EdgeNodeViewer.tsx & .css         (380 + 300 lines)
└── MetricsDisplay.tsx & .css         (400 + 300 lines)

vtp-frontend/src/services/
└── g5Service.ts                      (700+ lines)
```

## 🚀 Quick Start

### Import Components
```typescript
import NetworkStatus from '@/components/NetworkStatus';
import QualitySelector from '@/components/QualitySelector';
import EdgeNodeViewer from '@/components/EdgeNodeViewer';
import MetricsDisplay from '@/components/MetricsDisplay';
```

### Use in Dashboard (Already Done!)
```typescript
<NetworkStatus refreshInterval={5000} />
<MetricsDisplay refreshInterval={5000} />
<QualitySelector onProfileChanged={(p) => console.log(p)} />
<EdgeNodeViewer refreshInterval={10000} />
```

## 📊 Component Overview

| Component | Purpose | Key Features |
|-----------|---------|--------------|
| **NetworkStatus** | Display current network | Type badge, quality %, 5G indicator, health status |
| **QualitySelector** | Manage video profiles | 5 profiles, comparison table, AI recommendations |
| **EdgeNodeViewer** | Show edge nodes | Sorting, closest node highlight, capacity bars, stats |
| **MetricsDisplay** | Real-time metrics | Session metrics, global metrics, trend charts |

## 🔌 API Endpoints (14 Total)

```
GET    /api/5g/status                    # Adapter status
GET    /api/5g/health                    # Health check
POST   /api/5g/network/detect            # Detect network
GET    /api/5g/network/current           # Current network info
GET    /api/5g/network/quality           # Quality score (0-100)
GET    /api/5g/network/5g-available      # 5G available? (bool)
GET    /api/5g/quality/profiles          # All profiles
GET    /api/5g/quality/current           # Current profile
POST   /api/5g/quality/set               # Set profile
GET    /api/5g/edge/nodes                # All edge nodes
GET    /api/5g/edge/closest              # Closest node
GET    /api/5g/metrics/session           # Session metrics
GET    /api/5g/metrics/global            # Global metrics
POST   /api/5g/metrics/record            # Record metric
POST   /api/5g/session/start             # Start session
POST   /api/5g/session/end               # End session
```

## 💻 g5Service Methods

```typescript
// Network Info
g5Service.getCurrentNetwork()          // NetworkInfo
g5Service.getNetworkQuality()          // number (0-100)
g5Service.is5GAvailable()              // boolean
g5Service.detectNetwork()              // DetectionResult

// Quality Profiles
g5Service.getQualityProfiles()         // QualityProfile[]
g5Service.getCurrentQualityProfile()   // QualityProfile
g5Service.setQualityProfile(id)        // QualityProfile

// Edge Nodes
g5Service.getAvailableEdgeNodes()      // EdgeNode[]
g5Service.getClosestEdgeNode()         // EdgeNode

// Metrics
g5Service.getSessionMetrics()          // SessionMetrics
g5Service.getGlobalMetrics()           // GlobalMetrics
g5Service.recordMetric(data)           // void

// Sessions
g5Service.startSession()               // Session
g5Service.endSession(id)               // void

// Status
g5Service.getStatus()                  // AdapterStatus
g5Service.healthCheck()                // HealthResponse
```

## 🎨 Component Props

### NetworkStatus
```typescript
interface Props {
  refreshInterval?: number;           // Default: 5000ms
  onStatusChange?: (status: NetworkInfo) => void;
}
```

### QualitySelector
```typescript
interface Props {
  onProfileChanged?: (profile: QualityProfile) => void;
  refreshInterval?: number;
}
```

### EdgeNodeViewer
```typescript
interface Props {
  refreshInterval?: number;           // Default: 10000ms
  onNodeSelected?: (node: EdgeNode) => void;
}
```

### MetricsDisplay
```typescript
interface Props {
  refreshInterval?: number;           // Default: 5000ms
  onMetricsUpdate?: (session: SessionMetrics, global: GlobalMetrics) => void;
}
```

## 🎯 Key Types

```typescript
// Network
interface NetworkInfo {
  type: '5G' | '4G' | 'LTE' | 'WiFi';
  signalStrength: number;              // 0-100
  bandwidth: number;                   // Kbps
  latency: number;                     // ms
}

// Edge Node
interface EdgeNode {
  id: string;
  name: string;
  region: string;
  latency: number;                     // ms
  capacity: number;                    // GB
  available: number;                   // GB
  status: 'online' | 'offline' | 'degraded' | 'maintenance';
}

// Quality Profile
interface QualityProfile {
  id: string;
  name: string;
  resolution: string;                  // e.g., "1080p"
  codec: string;                       // e.g., "H.264"
  minBandwidth: number;                // Kbps
  maxLatency: number;                  // ms
}

// Session Metrics
interface SessionMetrics {
  latency: number;
  bandwidth: number;
  packetLoss: number;
  qualityLevel: number;                // 0-100
  codec: string;
  resolution: string;
}

// Global Metrics
interface GlobalMetrics {
  averageLatency: number;
  totalBandwidth: number;
  networkQuality: number;              // 0-100
  activeUsers: number;
  peakBandwidth: number;
  uptime: number;                      // seconds
}
```

## 🎨 Styling

All components use:
- **Dark Theme:** Gradient backgrounds (#1e1e1e → #2d2d2d)
- **Colors:**
  - Green: #66ff66 (good)
  - Orange: #ffaa00 (warning)
  - Red: #ff6666 (poor)
  - Purple: #667eea (primary)
- **Responsive:** Mobile-first design
- **Animations:** Smooth transitions, pulsing effects

## ⚙️ Configuration

### Environment Variables
```env
# .env.local
NEXT_PUBLIC_5G_API_BASE_URL=http://localhost:8080/api/5g
```

### Refresh Intervals
```typescript
NetworkStatus:   5000ms  (5 seconds)
MetricsDisplay:  5000ms  (5 seconds)
EdgeNodeViewer: 10000ms  (10 seconds)
QualitySelector: On demand (no auto-refresh)
```

## 📈 Quality Profiles

The 5 built-in quality profiles:

| Profile | Resolution | Codec | Min BW | Max Latency |
|---------|-----------|-------|--------|-------------|
| Ultra HD | 4K (3840×2160) | H.265 | 8000 Kbps | 20ms |
| HD | 1080p (1920×1080) | H.264 | 5000 Kbps | 40ms |
| Standard | 720p (1280×720) | H.264 | 2500 Kbps | 60ms |
| Medium | 480p (854×480) | H.264 | 1500 Kbps | 100ms |
| Low | 360p (640×360) | H.264 | 500 Kbps | 150ms |

## 🔍 Debugging

### Check Console Errors
1. Open browser DevTools (F12)
2. Go to Console tab
3. Look for errors from g5Service calls

### Check Network Requests
1. Go to Network tab
2. Filter by XHR requests
3. Check `/api/5g/*` endpoints
4. Verify response status (200)

### Verify Backend Connection
```typescript
// In browser console
import { g5Service } from '@/services/g5Service';
await g5Service.healthCheck();
```

## 📁 File Locations

- Dashboard Page: `vtp-frontend/src/app/dashboard/page.tsx`
- Components: `vtp-frontend/src/components/`
- Service: `vtp-frontend/src/services/g5Service.ts`
- Documentation: `PHASE_5G_DAY7_COMPLETION_REPORT.md`

## ✅ Verification Checklist

- [ ] Go backend running on port 8080
- [ ] Frontend dev server running (`npm run dev`)
- [ ] Dashboard loads at http://localhost:3000/dashboard
- [ ] 5G components appear at bottom of dashboard
- [ ] NetworkStatus shows network type and quality
- [ ] MetricsDisplay shows current metrics
- [ ] QualitySelector shows 5 profiles
- [ ] EdgeNodeViewer shows available nodes
- [ ] No errors in browser console
- [ ] API calls visible in Network tab

## 🔗 Related Documentation

- **Phase 5G Overview:** `README.md`
- **Phase 5G Day 6 Results:** `PHASE_5G_DAY6_*` files
- **Detailed Report:** `PHASE_5G_DAY7_COMPLETION_REPORT.md`
- **Backend API:** `pkg/g5/` Go source code

## 📞 Common Issues

### Components not appearing
- Ensure imports are correct
- Check browser console for errors
- Verify g5Service is imported

### API calls failing
- Check Go backend is running on port 8080
- Verify `.env.local` has correct API base URL
- Check Network tab in DevTools for failed requests

### Styling looks wrong
- Clear browser cache (Ctrl+Shift+Delete)
- Restart dev server (`npm run dev`)
- Check CSS files are in correct location

### Metrics not updating
- Check auto-refresh interval setting
- Verify backend returns valid data
- Check console for API errors

---

**Phase 5G Day 7: COMPLETE ✅**

All 4 components + service integration successfully built and integrated into dashboard.
