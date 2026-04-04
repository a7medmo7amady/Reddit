workspace "Reddit Clone" {
    !identifiers hierarchical
    model {
        U = person "user"
        A = person "admin"
        R = softwareSystem "Reddit" {
        SearchService = container "Search service" {
            SearchApp = component "Search app"
            SearchMongo = component "Search MongoDB" {
                tags "Database"
            }
            SearchApp -> SearchMongo "Reads/writes search index documents"
        }
        UserService = container "User service" {
            UserApp = component "User app"
            UserPostgres = component "User Postgres" {
                tags "Database"
            }
            UserApp -> UserPostgres "Reads/writes users, communities, follows, relationships"
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
