package com.reddit.clone.model;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
import java.time.LocalDateTime;

@Document(collection = "moderation_logs")
public class ModerationLog {

    @Id
    private String id;

    private String action;
    private Long moderatorId;
    private String moderatorUsername;
    private Long targetUserId;
    private String targetUsername;
    private String reason;
    private LocalDateTime timestamp;

    public ModerationLog() {}

    public ModerationLog(String action, Long moderatorId, String moderatorUsername, 
                         Long targetUserId, String targetUsername, String reason) {
        this.action = action;
        this.moderatorId = moderatorId;
        this.moderatorUsername = moderatorUsername;
        this.targetUserId = targetUserId;
        this.targetUsername = targetUsername;
        this.reason = reason;
        this.timestamp = LocalDateTime.now();
    }

    // Getters and Setters
    public String getId() { return id; }
    public String getAction() { return action; }
    public Long getModeratorId() { return moderatorId; }
    public String getModeratorUsername() { return moderatorUsername; }
    public Long getTargetUserId() { return targetUserId; }
    public String getTargetUsername() { return targetUsername; }
    public String getReason() { return reason; }
    public LocalDateTime getTimestamp() { return timestamp; }
}
