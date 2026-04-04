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
        UploadService = container "Upload service" {
            UploadApp = component "Upload app"
            UploadMongo = component "Upload MongoDB" {
                tags "Database"
            }
            UploadApp -> UploadMongo "Reads/writes upload metadata and draft documents"
        }
        ChatService = container "Chat service" {
            ChatApp = component "Chat app"
            ChatMongo = component "Chat MongoDB" {
                tags "Database"
            }
            ChatApp -> ChatMongo "Reads/writes chat messages"
        }
        FeedService = container "Feed service" {
            FeedApp = component "Feed app"
            FeedMongo = component "Feed MongoDB" {
                tags "Database"
            }
            FeedApp -> FeedMongo "Reads/writes posts, comments, votes"
        }
        VideoService = container "Video service" {
            VideoApp = component "Video app"
            VideoMongo = component "Video MongoDB" {
                tags "Database"
            }
            VideoApp -> VideoMongo "Reads/writes video metadata"
        }
        NotificationService = container "NotificationService" {
            NotificationApp = component "Notification app"
            NotificationMongo = component "Notification MongoDB" {
                tags "Database"
            }
            NotificationApp -> NotificationMongo "Reads/writes activity, audit, mod logs"
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
