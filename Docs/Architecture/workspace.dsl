workspace "Reddit Clone" {
    !identifiers hierarchical
    model {
        U = person "user"
        A = person "admin"
        R = softwareSystem "Reddit" {
        SearchService = container "Search service" {
            SearchController  = component "Search Controller" "Exposes GET /search?q=&type=&date=&community="
            SearchIndexer     = component "Search Indexer" "Indexes posts/comments on creation, removes on deletion."
            FilterHandler     = component "Filter Handler" "Applies content type (Text, Image, Video) and date range filters."
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
            RegistrationHandler = component "Registration Handler" "Validates username, email, and password. Sends verification email. Username is permanent."
            AuthHandler         = component "Auth Handler" "Password login issues JWT + refresh token. Rotates refresh token on every use. Detects token family reuse and invalidates session."
            OAuthHandler        = component "OAuth Handler" "Handles OAuth 2.0 login via Google. Issues internal JWT."
            SessionManager      = component "Session Manager" "Generates and signs JWTs. Rotates JWT weekly. Saves session token to user's browser via httpOnly cookie. Validates session on each request."
            ProfileHandler      = component "Profile Handler" "Serves public profile: avatar, banner, display name, karma, account age, bio, and links."
            AccountSettings     = component "Account Settings Handler" "Change email, password, display name. Enable/disable 2FA via TOTP. Manage notification preferences."
            FollowBlockHandler  = component "Follow & Block Handler" "Follow users for Following feed. Block hides content and prevents messages/mentions."
            ModerationHandler   = component "Moderation Handler" "Moderator actions: remove post/comment, ban user, set rules — logged in mod log. Admins can suspend or permanently ban accounts."
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
            postController     = component "Post Controller"      "Exposes REST: POST /posts, PATCH /posts/{id}, DELETE /posts/{id}. Post types: Text, Image, Link, Video."
            authClient         = component "Auth Client"          "Validates JWT and checks community ban status via UserService."
            imageProcessor     = component "Image Processor"      "Accepts JPEG/PNG/GIF/WebP. Strips EXIF metadata, converts to WebP, resizes to thumbnail 140px, preview 640px, full 1080px."
            videoIntakeHandler = component "Video Intake Handler" "Accepts MP4/MOV/WebM (H.264/H.265). Validates format and constraints then enqueues async HLS transcoding job."
            postRepository     = component "Post Repository"      "Creates, updates, soft-deletes posts. Stores edit history for text/link posts. Deleted posts show [deleted] placeholder."
            storageClient      = component "Storage Client"       "Uploads processed media to S3 staging bucket. Returns pre-signed CDN URL."
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
        ChatService = container "Chat Service" "Real-time text messaging for DMs and community chat rooms." "Node.js" {
            wsGateway         = component "WebSocket Gateway"       "Manages WebSocket connections. Shared with Notification Service. Handles reconnect and missed message fetch."
            dmHandler         = component "DM Handler"              "Manages 1-to-1 direct messages. Checks block status. Plain text only, max 2,000 chars. Users can delete their own messages."
            roomHandler       = component "Community Room Handler"  "One chat room per community. Members send messages; moderators can delete any message. Loads last 7 days of history."
            messageProcessor  = component "Message Processor"       "Validates content. Renders URLs as clickable links. Detects @mentions and emits mention event to Notification Service."
            chatInbox         = component "Chat Inbox"              "Lists all DM conversations and community rooms sorted by most recent message. Tracks unread count per conversation."
            moderationHandler = component "Chat Moderation Handler" "Handles message reports. Enforces block list. Allows moderators to delete messages in their community room."
            ChatMongo         = component "Chat MongoDB"            "Stores messages, conversation metadata, and unread counts." {
                tags "Database"
            }

            wsGateway         -> dmHandler          "Route DM message"
            wsGateway         -> roomHandler        "Route community room message"
            dmHandler         -> moderationHandler  "Check block status"
            dmHandler         -> messageProcessor   "Process message content"
            roomHandler       -> messageProcessor   "Process message content"
            messageProcessor  -> ChatMongo          "Persist message"
            messageProcessor  -> wsGateway          "Push to recipient(s)"
            chatInbox         -> ChatMongo          "Read conversations and unread counts"
            moderationHandler -> ChatMongo          "Delete message / enforce block"
        }
        FeedService = container "Feed Service" "Aggregates, ranks, and serves personalized and discovery feeds." "Go" {
            feedController   = component "Feed Controller"    "Exposes GET /feed/home, /feed/community/{id}, /feed/popular, /feed/all. Sort: hot, new, rising, controversial. Batches of 25 with cursor."
            authFeedBuilder  = component "Auth Feed Builder"  "Builds personalized Home Feed using ranking algorithm based on subscriptions, vote history, view history, and Not Interested signals."
            guestFeedBuilder = component "Guest Feed Builder" "Serves default curated feed for guests. Guests can navigate to Popular/All. Injects sign-up banner."
            communityFeed    = component "Community Feed"     "Scopes feed to a single community. Pins up to 2 posts at top. Returns community sidebar."
            discoveryFeed    = component "Discovery Feed"     "Popular: trending posts via Count-Min Sketch. All: pure chronological. Available to guests and authenticated users."
            votingHandler    = component "Voting Handler"     "Upvote/downvote once per post/comment. Toggle same vote to remove. Optimistic client update reconciled server-side. Updates karma."
            personalization  = component "Personalization"    "Records viewed post IDs, vote history, community prefs. Applies hide-community and Not Interested signals."
            kafkaConsumer    = component "Kafka Consumer"     "Subscribes to post.created and post.deleted to update feed indexes."
            FeedMongo        = component "Feed MongoDB"       "Stores posts, vote records, personalization signals, and view history." {
                tags "Database"
            }
            FeedRedis        = component "Feed Redis"         "Caches ranked feed results. TTL: 60s for Hot/Rising; real-time for New." {
                tags "Database"
            }

            feedController   -> authFeedBuilder   "Authenticated user request"
            feedController   -> guestFeedBuilder  "Guest default feed request"
            guestFeedBuilder -> discoveryFeed     "Guest navigates Popular / All"
            feedController   -> communityFeed     "Community-scoped request"
            feedController   -> discoveryFeed     "Popular / All request"
            authFeedBuilder  -> personalization   "Apply user signals"
            authFeedBuilder  -> FeedRedis         "Read/write cached feed"
            authFeedBuilder  -> FeedMongo         "Query posts when cache miss"
            communityFeed    -> FeedRedis         "Read/write cached feed"
            communityFeed    -> FeedMongo         "Query posts when cache miss"
            discoveryFeed    -> FeedRedis         "Read/write cached feed"
            discoveryFeed    -> FeedMongo         "Query posts when cache miss"
            votingHandler    -> FeedMongo         "Write vote record"
            votingHandler    -> FeedRedis         "Invalidate affected feed cache"
            personalization  -> FeedMongo         "Read/write user signals"
            kafkaConsumer    -> FeedMongo         "Index new / remove deleted posts"
            kafkaConsumer    -> FeedRedis         "Invalidate cache on new post"
        }
        VideoService = container "Video Service" "Handles async transcoding, HLS delivery, and in-feed video playback." "Go" {
            kafkaConsumer       = component "Kafka Consumer"        "Subscribes to video.uploaded event from Upload Service. Triggers transcoding pipeline."
            transcodingPipeline = component "Transcoding Pipeline"  "Transcodes video into HLS variants: 360p, 720p, 1080p. Audio: AAC 128kbps stereo. Segments: 6s .ts chunks + .m3u8 manifest."
            thumbnailExtractor  = component "Thumbnail Extractor"   "Extracts static thumbnail at t=3s as WebP. Generates 3-second animated preview for feed hover."
            manifestBuilder     = component "Manifest Builder"      "Builds .m3u8 master manifest listing all quality variants for adaptive bitrate streaming."
            statusTracker       = component "Status Tracker"        "Tracks transcoding job state (queued, processing, done, failed). Exposes GET /videos/{id}/status."
            storageClient       = component "Storage Client"        "Uploads .ts segments, .m3u8 manifests, thumbnail, and animated preview to S3."
            playbackHandler     = component "Playback Handler"      "Serves HLS stream URLs. Enforces single-video playback. Handles autoplay muted on scroll-into-view."
            VideoMongo          = component "Video MongoDB"         "Stores video metadata: status, manifest URL, thumbnail URL, animated preview URL, variants, duration." {
                tags "Database"
            }

            kafkaConsumer       -> transcodingPipeline "Trigger transcode job"
            transcodingPipeline -> thumbnailExtractor  "Extract thumbnail and animated preview"
            transcodingPipeline -> manifestBuilder     "Build HLS manifest"
            transcodingPipeline -> storageClient       "Upload .ts segments"
            thumbnailExtractor  -> storageClient       "Upload WebP thumbnail and animated preview"
            manifestBuilder     -> storageClient       "Upload .m3u8 manifest"
            transcodingPipeline -> statusTracker       "Update job status"
            transcodingPipeline -> VideoMongo          "Write video metadata on completion"
            statusTracker       -> VideoMongo          "Read/write job status"
            playbackHandler     -> VideoMongo          "Read manifest and thumbnail URLs"
            playbackHandler     -> storageClient       "Serve CDN URLs to client"
        }
        NotificationService = container "Notification Service" "Delivers in-app and email notifications for platform events." "Node.js" {
            kafkaConsumer      = component "Kafka Consumer"       "Subscribes to post.reply, comment.reply, mention, dm.received, post.removed events."
            notificationRouter = component "Notification Router"  "Decides delivery channel per event and user preference. In-app for all events; email only for DM when user is offline."
            inAppHandler       = component "In-App Handler"       "Creates notification records. Pushes to client via WebSocket. If offline, queues in Redis for delivery on next connection."
            emailHandler       = component "Email Handler"        "Sends transactional email for new DM when user is offline. Uses Nodemailer + SMTP."
            preferenceChecker  = component "Preference Checker"   "Checks user notification preferences before delivery. Respects global in-app on/off toggle."
            retentionCleaner   = component "Retention Cleaner"    "Deletes notification records older than 30 days per user."
            NotificationMongo  = component "Notification MongoDB" "Stores notification records (type, summary, timestamp, read status)." {
                tags "Database"
            }
            NotificationRedis  = component "Notification Redis"  "Queues notifications for offline users. Delivered on next WebSocket connection." {
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
        VideoService  -> S3 "Stores video objects"
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
