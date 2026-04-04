workspace "Reddit Clone" {
    !identifiers hierarchical
    model {
        U = person "user"
        A = person "admin"
        R = softwareSystem "Reddit" {
        RDB = container "Postgres"
        NRDB = container "Mongo"
        SearchService = container "Search service"
        UserService = container "User service"
        UploadService = container "Upload service"
        ChatService = container "Chat service"
        FeedService = container "Feed service"
        VideoService = container "Video service"
        S3 = container "S3 Bucket"
        RDBVideo = container "Video Postgres" {
            tags "Database"
        }
        NotificationService = container "NotificationService"
        }
        U -> R "Uses"
        A -> R "Monitors"
    }

    views {
            systemContext R "Context" {
            include *
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
