workspace "Reddit Clone" {
    !identifiers hierarchical
    model {
        U = person "user"
        A = person "admin"
        R = softwareSystem "Reddit" {
        SearchService = container "Search service" {
            SearchController  = component "Search Controller" "Exposes GET /search?q=&type=&date=&community="
            SearchIndexer     = component "Search Indexer" "Indexes posts/comments on creation, removes on deletion."
            FilterHandler     = component "Filter Handler" "Applies content type (Text, Image, Video) and date range (Week, Month, All Time) filters."
            ScopedSearch      = component "Scoped Search Handler" "Narrows results to a specific community when community_id is provided."
            PaginationHandler = component "Pagination Handler" "Returns 20 results per page with a cursor for Load More."
            SearchMongo       = component "Search MongoDB" "Stores search index documents." {
                tags "Database"
            }

            SearchController  -> FilterHandler     "Applies filters"
            SearchController  -> ScopedSearch      "Scopes to community if provided"
            SearchController  -> PaginationHandler "Paginates results"
            SearchController  -> SearchMongo       "Queries index"
            SearchIndexer     -> SearchMongo       "Indexes / removes documents"
        }
  UserService = container "User service" {
    RegistrationHandler = component "Registration Handler" "Validates username (3-20 chars), email, password (min 8 chars). Sends verification email. Username is permanent."
    AuthHandler         = component "Auth Handler" "Password login issues JWT (15 min) + refresh token (30 days, httpOnly cookie). Rotates refresh token on every use. Detects token family reuse and invalidates session."
    OAuthHandler        = component "OAuth Handler" "Handles OAuth 2.0 login via Google. Delegates to external provider, then issues internal JWT."
    SessionManager      = component "Session Manager" "Generates and signs JWTs. Rotates JWT weekly. Saves session token to user's browser via httpOnly cookie. Validates session on each request."
    ProfileHandler      = component "Profile Handler" "Serves public profile: avatar, banner, display name, karma, account age, bio, links (up to 3). Tabs: Posts, Comments, Saved, Hidden, Upvoted."
    AccountSettings     = component "Account Settings Handler" "Change email (re-verification required), password, display name. Enable/disable 2FA via TOTP. Manage notification preferences."
    FollowBlockHandler  = component "Follow & Block Handler" "Follow users for Following feed. Block hides content and prevents messages/mentions. Blocked list is private."
    ModerationHandler   = component "Moderation Handler" "Moderator actions: remove post/comment, ban user, set rules — all logged in mod log. Admins can suspend or permanently ban accounts."
    UserPostgres        = component "User Postgres" "Stores users, sessions, follows, blocks, roles, community memberships." {
        tags "Database"
    }
    UserMongo           = component "User MongoDB" "Stores posts, comments and mod logs." {
        tags "Database"
    }

    RegistrationHandler -> UserPostgres "Reads/writes user accounts"
    AuthHandler         -> UserPostgres "Reads/writes sessions and tokens"
    OAuthHandler        -> UserPostgres "Upserts OAuth-linked accounts"
    ProfileHandler      -> UserPostgres "Reads user data and community info"
    ProfileHandler      -> UserMongo    "Reads posts and comments for profile tabs"
    AccountSettings     -> UserPostgres "Updates account settings and 2FA"
    FollowBlockHandler  -> UserPostgres "Reads/writes follows and blocks"
    ModerationHandler   -> UserPostgres "Reads/writes roles"
    ModerationHandler   -> UserMongo    "Writes mod log entries"
}
        UploadService = container "Upload Service" "Handles post creation, media processing, and CDN delivery." "Node.js" {
            postController     = component "Post Controller"      "Exposes REST: POST /posts, PATCH /posts/{id}, DELETE /posts/{id}. Enforces required fields: title (1-300 chars), community. Optional: flair, NSFW, Spoiler, OC badge. Post types: Text, Image, Link, Video."
            authClient         = component "Auth Client"          "Validates JWT and checks community ban status via UserService before allowing post creation."
            imageProcessor     = component "Image Processor"      "Accepts JPEG/PNG/GIF/WebP up to 20MB, max 5 images per gallery. Strips EXIF metadata, converts to WebP, resizes to 3 breakpoints: thumbnail 140px, preview 640px, full 1080px. Returns pre-signed CDN URL. Target: response within 3s for files ≤5MB."
            videoIntakeHandler = component "Video Intake Handler" "Accepts MP4/MOV/WebM (H.264/H.265), max 100MB, max duration 5 minutes. Validates format and constraints then enqueues async HLS transcoding job (360p, 720p, 1080p variants)."
            postRepository     = component "Post Repository"      "Creates, updates, soft-deletes posts. Stores edit history (text/link posts only — media cannot be replaced). Deleted posts show [deleted] placeholder; content purged from storage within 24h."
            storageClient      = component "Storage Client"       "Uploads processed media to S3 staging bucket. Returns pre-signed CDN URL with long-lived cache headers."
            kafkaProducer      = component "Kafka Producer"       "Emits post.created and post.deleted events to the message broker."
            UploadMongo        = component "Upload MongoDB"       "Stores post metadata, edit history, and draft documents." {
                tags "Database"
            }

            postController     -> authClient         "Validates JWT + ban status"
            postController     -> imageProcessor     "On image post"
            postController     -> videoIntakeHandler "On video post"
            postController     -> postRepository     "Save post metadata (all post types)"
            imageProcessor     -> storageClient      "Upload processed WebP"
            videoIntakeHandler -> storageClient      "Upload raw video to staging bucket"
            videoIntakeHandler -> kafkaProducer      "Emit video.uploaded event"
            postRepository     -> UploadMongo        "Reads/writes post metadata and edit history"
            postRepository     -> kafkaProducer      "Emit post.created / post.deleted"
        }
        ChatService = container "Chat service" {
            ChatApp = component "Chat app"
            ChatMongo = component "Chat MongoDB" {
                tags "Database"
            }
            ChatApp -> ChatMongo "Reads/writes chat messages"
        }
        FeedService = container "Feed Service" "Aggregates, ranks, and serves personalized and discovery feeds." "Go" {
            feedController    = component "Feed Controller"     "Exposes REST: GET /feed/home, /feed/community/{id}, /feed/popular, /feed/all. Supports sort params: hot, new, rising, controversial. Returns batches of 25 posts with cursor for infinite scroll."
            authFeedBuilder   = component "Auth Feed Builder"  "Builds personalized Home Feed for authenticated users using a ranking algorithm based on subscription history, vote history, view history, and Not Interested signals. Excludes hidden communities."
            guestFeedBuilder  = component "Guest Feed Builder" "Serves default curated feed for guests. Guests can also navigate to Popular/All discovery feeds. Injects sticky sign-up banner into response."
            communityFeed     = component "Community Feed"     "Scopes feed to a single community. Pins up to 2 pinned posts at top regardless of sort. Returns community sidebar: description, rules, moderators, member count, creation date."
            discoveryFeed     = component "Discovery Feed"     "Popular: trending posts detected using a Count-Min Sketch to estimate post engagement frequency with low memory overhead. All: pure chronological feed of all new posts. Both available to guests and authenticated users."
            votingHandler     = component "Voting Handler"     "Authenticated users upvote/downvote once per post/comment. Toggle same vote to remove. Optimistic client update reconciled server-side. Updates karma in real time."
            personalization   = component "Personalization"    "Records viewed post IDs, vote history, community prefs per user. Applies hide-community and Not Interested down-ranking signals to Home Feed."
            kafkaConsumer     = component "Kafka Consumer"     "Subscribes to post.created and post.deleted events to update feed indexes."
            FeedMongo         = component "Feed MongoDB"       "Stores posts, vote records, personalization signals, and view history." {
                tags "Database"
            }
            FeedRedis         = component "Feed Redis"         "Caches ranked feed results. TTL: 60s for Hot/Rising; no cache for New (real-time). Supports ≥10,000 concurrent requests/sec at peak." {
                tags "Database"
            }

            feedController  -> authFeedBuilder  "Authenticated user request"
            feedController  -> guestFeedBuilder "Guest default feed request"
            guestFeedBuilder -> discoveryFeed  "Guest navigates Popular / All"
            feedController  -> communityFeed    "Community-scoped request"
            feedController  -> discoveryFeed    "Popular / All request"
            authFeedBuilder -> personalization  "Apply user signals"
            authFeedBuilder -> FeedRedis        "Read/write cached feed"
            authFeedBuilder -> FeedMongo        "Query posts when cache miss"
            communityFeed   -> FeedRedis        "Read/write cached feed"
            communityFeed   -> FeedMongo        "Query posts when cache miss"
            discoveryFeed   -> FeedRedis        "Read/write cached feed"
            discoveryFeed   -> FeedMongo        "Query posts when cache miss"
            votingHandler   -> FeedMongo        "Write vote record"
            votingHandler   -> FeedRedis        "Invalidate affected feed cache"
            personalization -> FeedMongo        "Read/write user signals"
            kafkaConsumer   -> FeedMongo        "Index new / remove deleted posts"
            kafkaConsumer   -> FeedRedis        "Invalidate cache on new post"
        }
        VideoService = container "Video service" {
            VideoApp = component "Video app"
            VideoMongo = component "Video MongoDB" {
                tags "Database"
            }
            VideoApp -> VideoMongo "Reads/writes video metadata"
        }
        NotificationService = container "Notification Service" "Delivers in-app and email notifications for platform events." "Node.js" {
            kafkaConsumer      = component "Kafka Consumer"        "Subscribes to events: post.reply, comment.reply, mention, dm.received, post.removed. Triggers appropriate notification flow."
            notificationRouter = component "Notification Router"   "Decides delivery channel per event and user preference. In-app for all events; email only for DM when user is offline."
            inAppHandler       = component "In-App Handler"        "Creates notification records. Pushes to client via WebSocket (shared with Chat). If user offline, queues in Redis for delivery on next connection. p95 delivery ≤2s for connected users."
            emailHandler       = component "Email Handler"         "Sends transactional email for new DM when user is offline. Uses Nodemailer + SMTP (Mailtrap for dev, SES for prod)."
            preferenceChecker  = component "Preference Checker"    "Checks user notification preferences before delivery. Respects global in-app on/off toggle from account settings."
            retentionCleaner   = component "Retention Cleaner"     "Deletes notification records older than 30 days per user."
            NotificationMongo  = component "Notification MongoDB"  "Stores notification records (type, summary, timestamp, read status). Retained 30 days." {
                tags "Database"
            }
            NotificationRedis  = component "Notification Redis"    "Queues notifications for offline users. Delivered on next WebSocket connection." {
                tags "Database"
            }

            kafkaConsumer      -> notificationRouter  "Routes event"
            notificationRouter -> preferenceChecker   "Check user preferences"
            notificationRouter -> inAppHandler        "Trigger in-app notification"
            notificationRouter -> emailHandler        "Trigger email (DM + offline only)"
            inAppHandler       -> NotificationMongo   "Write notification record"
            inAppHandler       -> NotificationRedis   "Queue if user offline"
            retentionCleaner   -> NotificationMongo   "Delete records older than 30 days"
        }
        S3 = container "S3 Bucket"

        UploadService -> S3 "Stores media objects"
        VideoService -> S3 "Stores video objects"
        }
        U -> R "Uses"
        A -> R "Monitors"
    }

    views {
            systemContext R "Context" {
            include *
            autolayout lr 300 300
        }
        container R "Containers" {
            include *
            autolayout lr 300 300
        }
        component R.SearchService Search_Components {
            include *
            autolayout lr 300 300
        }
        component R.UserService User_Components {
            include *
            autolayout lr 300 300
        }
        component R.UploadService Upload_Components {
            include *
            autolayout lr 300 300
        }
        component R.ChatService Chat_Components {
            include *
            autolayout lr 300 300
        }
        component R.FeedService Feed_Components {
            include *
            autolayout lr 300 300
        }
        component R.VideoService Video_Components {
            include *
            autolayout lr 300 300
        }
        component R.NotificationService Notification_Components {
            include *
            autolayout lr 300 300
        }
        
        styles {
            element "Element" {
                color #55aa55
                stroke #55aa55
                strokeWidth 7
                shape roundedbox
            }
            element "Person" {
                shape person
            }
            element "Database" {
                shape cylinder
            }
            element "Boundary" {
                strokeWidth 5
            }
            relationship "Relationship" {
                thickness 4
            }
        }
    }

}
