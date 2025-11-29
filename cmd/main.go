package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/yourusername/vtp-platform/pkg/auth"
	"github.com/yourusername/vtp-platform/pkg/course"
	"github.com/yourusername/vtp-platform/pkg/db"
	"github.com/yourusername/vtp-platform/pkg/instructor"
	"github.com/yourusername/vtp-platform/pkg/recording"
	"github.com/yourusername/vtp-platform/pkg/signalling"
	"github.com/yourusername/vtp-platform/pkg/streaming"
)

func getStorageDir() string {
	storageDir := os.Getenv("RECORDINGS_DIR")
	if storageDir == "" {
		storageDir = filepath.Join(os.TempDir(), "vtp-recordings")
	}

	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		log.Fatalf("❌ Failed to create recordings directory (%s): %v", storageDir, err)
	}

	log.Printf("      ✓ Recording directory: %s", storageDir)
	return storageDir
}

func getFFmpegPath() string {
	if envPath := os.Getenv("FFMPEG_PATH"); envPath != "" {
		if _, statErr := os.Stat(envPath); statErr == nil {
			log.Printf("      ✓ Using FFmpeg from FFMPEG_PATH: %s", envPath)
			return envPath
		} else {
			log.Printf("⚠ FFMPEG_PATH is set but invalid (%s): %v", envPath, statErr)
		}
	}

	if path, err := exec.LookPath("ffmpeg"); err == nil {
		log.Printf("      ✓ Found ffmpeg on PATH: %s", path)
		return path
	}

	log.Println("⚠ FFmpeg not found (transcoding disabled). Install FFmpeg or set FFMPEG_PATH.")
	return ""
}

func main() {
	// Load environment variables from .env file
	_ = godotenv.Load()

	// Log startup
	log.Println("═══════════════════════════════════════════════════════════════")
	log.Println("  VTP Platform - Educational Live Video Streaming System")
	log.Println("═══════════════════════════════════════════════════════════════")

	// 1. Initialize Database Connection
	log.Println("\n[1/5] Initializing database connection...")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Try common Railway and local defaults
		if railwayDBURL := os.Getenv("RAILWAY_DATABASE_URL"); railwayDBURL != "" {
			dbURL = railwayDBURL
			log.Println("      Using Railway DATABASE_URL")
		} else {
			dbURL = "postgres://postgres:postgres@localhost:5432/vtp_db?sslmode=disable"
			log.Println("      Using default local database URL")
		}
	}

	database, err := db.NewDatabase(dbURL)
	if err != nil {
		log.Printf("⚠ Database connection failed: %v", err)
		log.Println("      ⚠ Starting without database (recordings/streaming disabled)")
		database = nil
	} else {
		defer database.Close()
		log.Println("      ✓ Database connected")

		// 2. Run Database Migrations
		log.Println("\n[2/5] Running database migrations...")
		if err := database.RunMigrations(); err != nil {
			log.Printf("⚠ Migration failed: %v", err)
		} else {
			log.Println("      ✓ Migrations completed")
		}
	}

	// 3. Initialize Auth Services
	log.Println("\n[3/5] Initializing authentication services...")

	// Get JWT configuration
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Use default for development; MUST be set in production
		jwtSecret = "vtp-dev-default-secret-change-in-production"
		log.Println("⚠ JWT_SECRET not set, using development default (set JWT_SECRET in production!)")
	}

	jwtAccessHours := 24
	if val := os.Getenv("JWT_EXPIRY_HOURS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			jwtAccessHours = parsed
		}
	}

	jwtRefreshHours := 168
	if val := os.Getenv("JWT_REFRESH_EXPIRY_HOURS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			jwtRefreshHours = parsed
		}
	}

	// Initialize auth services
	tokenService := auth.NewTokenService(jwtSecret, jwtAccessHours, jwtRefreshHours)
	passwordService := auth.NewPasswordService(12)
	var userStore *auth.UserStore
	var authHandler *auth.AuthHandler
	var authMiddleware *auth.AuthMiddleware
	
	if database != nil {
		userStore = auth.NewUserStore(database.Conn(), passwordService)
		authHandler = auth.NewAuthHandler(userStore, tokenService, passwordService)
		authMiddleware = auth.NewAuthMiddleware(tokenService)
	} else {
		// Create dummy handlers when no database
		authHandler = auth.NewAuthHandler(nil, tokenService, passwordService)
		authMiddleware = auth.NewAuthMiddleware(tokenService)
	}

	log.Printf("      ✓ Token service (access: %dh, refresh: %dh)", jwtAccessHours, jwtRefreshHours)
	log.Println("      ✓ Password service (bcrypt cost: 12)")
	if database != nil {
		log.Println("      ✓ User store")
	}
	log.Println("      ✓ Auth handlers")
	log.Println("      ✓ Auth middleware")

	// 3b. Initialize Signalling Server (Phase 1b)
	log.Println("\n[3b/5] Initializing WebRTC signalling server...")
	sigServer, err := signalling.NewSignallingServer()
	if err != nil {
		log.Fatalf("❌ Failed to initialize signalling server: %v", err)
	}
	sigAPIHandler := signalling.NewAPIHandler(sigServer, authMiddleware)
	log.Println("      ✓ Socket.IO server initialized")
	log.Println("      ✓ Room manager initialized")
	log.Println("      ✓ Signalling handlers registered")

	// Ensure recording directories exist and configure ffmpeg path
	storageDir := getStorageDir()
	ffmpegPath := getFFmpegPath()

	// 3c. Initialize Recording Service (Phase 2a) - only if database available
	var recordingHandlers *recording.RecordingHandlers
	var storageHandlers *recording.StorageHandlers
	var playbackHandlers *recording.PlaybackHandlers
	
	if database != nil {
		log.Println("\n[3c/5] Initializing recording service...")
		recordingService := recording.NewRecordingService(database.Conn(), log.New(os.Stderr, "[Recording] ", log.LstdFlags))
		recordingHandlers = recording.NewRecordingHandlers(recordingService, log.New(os.Stderr, "[RecordingAPI] ", log.LstdFlags))

		// Initialize storage backend (Phase 2a Day 3)
		storageBackend, err := recording.NewLocalStorageBackend(storageDir, log.New(os.Stderr, "[Storage] ", log.LstdFlags))
		if err != nil {
			log.Printf("⚠ Failed to initialize storage backend: %v", err)
		} else {
			storageManager := recording.NewStorageManager(storageBackend, database.Conn(), log.New(os.Stderr, "[StorageManager] ", log.LstdFlags))
			storageHandlers = recording.NewStorageHandlers(storageManager, recordingService, log.New(os.Stderr, "[StorageAPI] ", log.LstdFlags))

			// Initialize streaming manager (Phase 2a Day 4)
			streamingManager := recording.NewStreamingManager(storageManager, database.Conn(), log.New(os.Stderr, "[Streaming] ", log.LstdFlags), storageDir)
			playbackHandlers = recording.NewPlaybackHandlers(streamingManager, recordingService, log.New(os.Stderr, "[PlaybackAPI] ", log.LstdFlags))

			log.Println("      ✓ Recording service initialized")
			log.Println("      ✓ Recording handlers initialized")
			log.Println("      ✓ Storage backend initialized (local filesystem)")
			log.Println("      ✓ Storage handlers initialized")
			log.Println("      ✓ Streaming manager initialized (HLS/DASH)")
			log.Println("      ✓ Playback handlers initialized")
		}
	} else {
		log.Println("\n[3c/5] Skipping recording service (no database)")
	}

	// 3d. Initialize Course Service (Phase 3) - only if database available
	var courseHandlers *course.CourseHandlers
	
	if database != nil {
		log.Println("\n[3d/5] Initializing course management service...")
		courseService := course.NewCourseService(database.Conn(), log.New(os.Stderr, "[CourseService] ", log.LstdFlags))
		courseHandlers = course.NewCourseHandlers(courseService, log.New(os.Stderr, "[CourseAPI] ", log.LstdFlags))

		log.Println("      ✓ Course service initialized")
		log.Println("      ✓ Course handlers initialized")
	} else {
		log.Println("\n[3d/5] Skipping course service (no database)")
	}

	// 3d2. Initialize Instructor Service (Phase 3+) - only if database available
	var instructorHandlers *instructor.Handler
	
	if database != nil {
		log.Println("\n[3d2/7] Initializing instructor management service...")
		instructorRepo := instructor.NewRepository(database.Conn())
		instructorService := instructor.NewService(instructorRepo)
		instructorHandlers = instructor.NewHandler(instructorService)

		log.Println("      ✓ Instructor repository initialized")
		log.Println("      ✓ Instructor service initialized")
		log.Println("      ✓ Instructor handlers initialized")
	} else {
		log.Println("\n[3d2/7] Skipping instructor service (no database)")
	}

	// 3e. Initialize Adaptive Bitrate (ABR) Engine (Phase 2B)
	log.Println("\n[3e/5] Initializing adaptive bitrate (ABR) streaming engine...")
	abrConfig := streaming.ABRConfig{
		MinBitrate:    500,  // 500 kbps minimum
		MaxBitrate:    4000, // 4000 kbps maximum
		ThresholdUp:   1.5,  // Scale up when bandwidth is 1.5x current bitrate
		ThresholdDown: 0.5,  // Scale down when bandwidth is 0.5x current bitrate
		HistorySize:   10,   // Keep 10 recent segments for analysis
	}
	abrManager := streaming.NewAdaptiveBitrateManager(abrConfig)
	abrHandlers := streaming.NewABRHandlers(abrManager, log.New(os.Stderr, "[ABRAPI] ", log.LstdFlags))

	log.Println("      ✓ ABR manager initialized")
	log.Println("      ✓ ABR engine configured (500-4000 kbps range)")
	log.Println("      ✓ ABR handlers initialized")

	// 3f. Initialize Multi-Bitrate Transcoder (Phase 2B Day 2)
	log.Println("\n[3f/7] Initializing multi-bitrate transcoding system...")
	transcoder := streaming.NewMultiBitrateTranscoder(storageDir, ffmpegPath, 4, log.New(os.Stderr, "[Transcoder] ", log.LstdFlags))
	transcodingService := streaming.NewTranscodingService(transcoder, 2, log.New(os.Stderr, "[TranscodingService] ", log.LstdFlags))
	transcodingHandlers := streaming.NewTranscodingHandlers(transcodingService, log.New(os.Stderr, "[TranscodingAPI] ", log.LstdFlags))

	log.Println("      ✓ Multi-bitrate transcoder initialized")
	log.Println("      ✓ Transcoding service (2 worker threads)")
	log.Println("      ✓ Transcoding handlers registered")

	// 3g. Initialize Live Distribution Network (Phase 2B Day 3)
	log.Println("\n[3g/7] Initializing live distribution network...")
	distributionService := streaming.NewDistributionService(4, log.New(os.Stderr, "[Distribution] ", log.LstdFlags))
	distributionService.EnableCDN("https://cdn.example.com")
	distributionHandlers := streaming.NewDistributionHandlers(distributionService)

	log.Println("      ✓ Distribution service initialized (4 workers)")
	log.Println("      ✓ CDN integration enabled")
	log.Println("      ✓ Distribution handlers registered")

	// 4. Register HTTP Routes
	log.Println("\n[4/7] Registering HTTP routes...")

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","service":"vtp-platform","version":"1.0.0"}`)
	})
	log.Println("      ✓ GET /health")

	// Authentication endpoints (public)
	http.HandleFunc("/api/v1/auth/register", authHandler.RegisterHandler)
	log.Println("      ✓ POST /api/v1/auth/register")

	http.HandleFunc("/api/v1/auth/login", authHandler.LoginHandler)
	log.Println("      ✓ POST /api/v1/auth/login")

	http.HandleFunc("/api/v1/auth/refresh", authHandler.RefreshHandler)
	log.Println("      ✓ POST /api/v1/auth/refresh")

	// Protected endpoints (require authentication)
	http.Handle("/api/v1/auth/profile",
		authMiddleware.Middleware(
			http.HandlerFunc(authHandler.GetProfileHandler)))
	log.Println("      ✓ GET /api/v1/auth/profile (protected)")

	http.Handle("/api/v1/auth/change-password",
		authMiddleware.Middleware(
			http.HandlerFunc(authHandler.ChangePasswordHandler)))
	log.Println("      ✓ POST /api/v1/auth/change-password (protected)")

	// Signalling endpoints (WebRTC)
	http.Handle("/socket.io/", sigServer)
	log.Println("      ✓ WebSocket /socket.io/ (WebRTC signalling)")

	http.HandleFunc("/api/v1/signalling/health", sigAPIHandler.HealthHandler)
	log.Println("      ✓ GET /api/v1/signalling/health")

	http.HandleFunc("/api/v1/signalling/room/stats", sigAPIHandler.GetRoomStatsHandler)
	log.Println("      ✓ GET /api/v1/signalling/room/stats")

	http.HandleFunc("/api/v1/signalling/rooms/stats", sigAPIHandler.GetAllRoomStatsHandler)
	log.Println("      ✓ GET /api/v1/signalling/rooms/stats")

	http.HandleFunc("/api/v1/signalling/room/create", sigAPIHandler.CreateRoomHandler)
	log.Println("      ✓ POST /api/v1/signalling/room/create")

	http.HandleFunc("/api/v1/signalling/room/delete", sigAPIHandler.DeleteRoomHandler)
	log.Println("      ✓ DELETE /api/v1/signalling/room/delete")

	// Recording endpoints (Phase 2a) - only if database available
	if recordingHandlers != nil {
		recordingHandlers.RegisterRoutes(http.DefaultServeMux)
		log.Println("      ✓ POST /api/v1/recordings/start")
		log.Println("      ✓ POST /api/v1/recordings/{id}/stop")
		log.Println("      ✓ GET /api/v1/recordings")
		log.Println("      ✓ GET /api/v1/recordings/{id}")
		log.Println("      ✓ DELETE /api/v1/recordings/{id}")
	}

	// Storage/Download endpoints (Phase 2a Day 3) - only if database available
	if storageHandlers != nil {
		storageHandlers.RegisterStorageRoutes(http.DefaultServeMux)
		log.Println("      ✓ GET /api/v1/recordings/{id}/download")
		log.Println("      ✓ GET /api/v1/recordings/{id}/download-url")
		log.Println("      ✓ GET /api/v1/recordings/{id}/info")
	}

	// Streaming/Playback endpoints (Phase 2a Day 4) - only if database available
	if playbackHandlers != nil {
		playbackHandlers.RegisterPlaybackRoutes(http.DefaultServeMux)
		log.Println("      ✓ GET /api/v1/recordings/{id}/stream/playlist.m3u8")
		log.Println("      ✓ GET /api/v1/recordings/{id}/stream/*.ts")
		log.Println("      ✓ POST /api/v1/recordings/{id}/transcode")
		log.Println("      ✓ POST /api/v1/recordings/{id}/progress")
		log.Println("      ✓ GET /api/v1/recordings/{id}/thumbnail")
		log.Println("      ✓ GET /api/v1/recordings/{id}/analytics")
	}

	// Course management endpoints (Phase 3) - only if database available
	if courseHandlers != nil {
		courseHandlers.RegisterCourseRoutes(http.DefaultServeMux)
		log.Println("      ✓ POST /api/v1/courses")
		log.Println("      ✓ GET /api/v1/courses")
		log.Println("      ✓ GET /api/v1/courses/{id}")
		log.Println("      ✓ PUT /api/v1/courses/{id}")
		log.Println("      ✓ DELETE /api/v1/courses/{id}")
		log.Println("      ✓ POST /api/v1/courses/{id}/enroll")
		log.Println("      ✓ GET /api/v1/courses/{id}/enrollments")
		log.Println("      ✓ DELETE /api/v1/courses/{id}/enroll/{student_id}")
		log.Println("      ✓ POST /api/v1/courses/{id}/recordings")
		log.Println("      ✓ POST /api/v1/courses/{id}/recordings/{recording_id}/publish")
		log.Println("      ✓ POST /api/v1/courses/{id}/permissions")
		log.Println("      ✓ GET /api/v1/courses/{id}/permissions/{user_id}")
		log.Println("      ✓ GET /api/v1/courses/{id}/stats")
	}

	// Instructor management endpoints (Phase 3+) - only if database available
	if instructorHandlers != nil {
		http.HandleFunc("/api/v1/instructors", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				instructorHandlers.ListInstructors(w, r)
			case http.MethodPost:
				instructorHandlers.CreateInstructor(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		http.HandleFunc("/api/v1/instructors/", func(w http.ResponseWriter, r *http.Request) {
			// Extract ID from path
			if r.URL.Path == "/api/v1/instructors/" {
				http.Error(w, "Instructor ID required", http.StatusBadRequest)
				return
			}

			switch r.Method {
			case http.MethodGet:
				// Check for /availability endpoint
				if r.URL.Path[len(r.URL.Path)-12:] == "/availability" {
					instructorHandlers.GetAvailableSlots(w, r)
				} else {
					instructorHandlers.GetInstructor(w, r)
				}
			case http.MethodPut:
				instructorHandlers.UpdateInstructor(w, r)
			case http.MethodDelete:
				instructorHandlers.DeleteInstructor(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		log.Println("      ✓ POST /api/v1/instructors")
		log.Println("      ✓ GET /api/v1/instructors")
		log.Println("      ✓ GET /api/v1/instructors/{id}")
		log.Println("      ✓ PUT /api/v1/instructors/{id}")
		log.Println("      ✓ DELETE /api/v1/instructors/{id}")
		log.Println("      ✓ GET /api/v1/instructors/{id}/availability")
	}

	// Adaptive Bitrate (ABR) endpoints (Phase 2B)
	abrHandlers.RegisterABRRoutes(http.DefaultServeMux)
	log.Println("      ✓ POST /api/v1/recordings/{id}/abr/quality")
	log.Println("      ✓ GET /api/v1/recordings/{id}/abr/stats")
	log.Println("      ✓ POST /api/v1/recordings/{id}/abr/metrics")

	// Multi-Bitrate Transcoding endpoints (Phase 2B Day 2)
	transcodingHandlers.RegisterTranscodingRoutes(http.DefaultServeMux)
	log.Println("      ✓ POST /api/v1/recordings/{id}/transcode/quality")
	log.Println("      ✓ GET /api/v1/recordings/{id}/transcode/progress")
	log.Println("      ✓ POST /api/v1/recordings/{id}/transcode/cancel")
	log.Println("      ✓ GET /api/v1/recordings/{id}/stream/master.m3u8")

	// Live Distribution Network endpoints (Phase 2B Day 3)
	distributionHandlers.RegisterDistributionRoutes(http.DefaultServeMux)
	log.Println("      ✓ POST /api/v1/streams/start")
	log.Println("      ✓ POST /api/v1/streams/{id} (join viewer)")
	log.Println("      ✓ GET /api/v1/streams/{id} (statistics)")
	log.Println("      ✓ DELETE /api/v1/streams/{id} (leave viewer)")
	log.Println("      ✓ POST /api/v1/segments/deliver")
	log.Println("      ✓ POST /api/v1/viewers/adapt-quality")
	log.Println("      ✓ GET /api/v1/distribution/metrics")
	log.Println("      ✓ GET /api/v1/distribution/health")

	// 5. Start HTTP Server
	log.Println("\n[5/7] Starting HTTP server...")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverAddr := ":" + port
	log.Printf("      ✓ Listening on http://localhost:%s\n", port)

	// Display available endpoints
	log.Println("\n═══════════════════════════════════════════════════════════════")
	log.Println("  Available Endpoints (Phase 1a + 1b + 2a + 3 + 2B Days 1-3)")
	log.Println("═══════════════════════════════════════════════════════════════")

	log.Println("\n  PHASE 1a - Authentication (public):")
	log.Printf("    POST   http://localhost:%s/api/v1/auth/register\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/auth/login\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/auth/refresh\n", port)
	log.Printf("    GET    http://localhost:%s/health\n", port)

	log.Println("\n  PHASE 1a - Authentication (protected):")
	log.Printf("    GET    http://localhost:%s/api/v1/auth/profile\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/auth/change-password\n", port)

	log.Println("\n  PHASE 1b - WebRTC Signalling:")
	log.Printf("    WS     ws://localhost:%s/socket.io/ (WebSocket)\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/signalling/health\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/signalling/room/stats?room_id=ROOM_ID\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/signalling/rooms/stats\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/signalling/room/create\n", port)
	log.Printf("    DELETE http://localhost:%s/api/v1/signalling/room/delete?room_id=ROOM_ID\n", port)

	log.Println("\n  PHASE 2a - Recording (protected):")
	log.Printf("    POST   http://localhost:%s/api/v1/recordings/start\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/recordings/{id}/stop\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}\n", port)
	log.Printf("    DELETE http://localhost:%s/api/v1/recordings/{id}\n", port)

	log.Println("\n  PHASE 2a Day 3 - Storage & Download (protected):")
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/download\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/download-url\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/info\n", port)

	log.Println("\n  PHASE 2a Day 4 - Streaming & Playback (protected):")
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/stream/playlist.m3u8\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/stream/segment-*.ts\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/recordings/{id}/transcode?format=hls\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/recordings/{id}/progress\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/thumbnail\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/analytics\n", port)

	log.Println("\n  PHASE 3 - Course Management (protected):")
	log.Printf("    POST   http://localhost:%s/api/v1/courses\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/courses\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/courses/{id}\n", port)
	log.Printf("    PUT    http://localhost:%s/api/v1/courses/{id}\n", port)
	log.Printf("    DELETE http://localhost:%s/api/v1/courses/{id}\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/courses/{id}/enroll\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/courses/{id}/enrollments\n", port)
	log.Printf("    DELETE http://localhost:%s/api/v1/courses/{id}/enroll/{student_id}\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/courses/{id}/recordings\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/courses/{id}/recordings/{recording_id}/publish\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/courses/{id}/permissions\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/courses/{id}/permissions/{user_id}\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/courses/{id}/stats\n", port)

	log.Println("\n  PHASE 2B - Adaptive Bitrate Streaming (protected):")
	log.Printf("    POST   http://localhost:%s/api/v1/recordings/{id}/abr/quality\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/abr/stats\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/recordings/{id}/abr/metrics\n", port)

	log.Println("\n  PHASE 2B Day 2 - Multi-Bitrate Transcoding (protected):")
	log.Printf("    POST   http://localhost:%s/api/v1/recordings/{id}/transcode/quality\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/transcode/progress\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/recordings/{id}/transcode/cancel\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/recordings/{id}/stream/master.m3u8\n", port)

	log.Println("\n  PHASE 2B Day 3 - Live Distribution Network (protected):")
	log.Printf("    POST   http://localhost:%s/api/v1/streams/start\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/streams/{id} (join)\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/streams/{id} (stats)\n", port)
	log.Printf("    DELETE http://localhost:%s/api/v1/streams/{id}?viewer_id=VIEWER_ID (leave)\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/segments/deliver\n", port)
	log.Printf("    POST   http://localhost:%s/api/v1/viewers/adapt-quality\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/distribution/metrics\n", port)
	log.Printf("    GET    http://localhost:%s/api/v1/distribution/health\n", port)

	log.Println("\n  Example Authorization Header:")
	log.Println("    Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")

	log.Println("\n═══════════════════════════════════════════════════════════════")
	log.Println("  Status: ✓ Phase 1a Complete - Phase 1b Complete - Phase 2a Complete - Phase 3 Complete - Phase 2B Day 3 Ready")
	log.Println("═══════════════════════════════════════════════════════════════")

	// Start server
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
