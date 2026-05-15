package com.reddit.clone.model;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.index.Indexed;
import org.springframework.data.mongodb.core.mapping.Document;
import java.time.LocalDateTime;

@Document(collection = "communities")
public class CommunityDocument {

    @Id
    private String id;

    private Long postgresId;

    @Indexed(unique = true)
    private String name;

    private String description;
    private Long creatorId;
    private String creatorUsername;
    private int memberCount;
    private LocalDateTime createdAt;

    public CommunityDocument() {}

    public CommunityDocument(Long postgresId, String name, String description,
                              Long creatorId, String creatorUsername) {
        this.postgresId = postgresId;
        this.name = name;
        this.description = description;
        this.creatorId = creatorId;
        this.creatorUsername = creatorUsername;
        this.memberCount = 1;
        this.createdAt = LocalDateTime.now();
    }

    public String getId() { return id; }
    public Long getPostgresId() { return postgresId; }
    public String getName() { return name; }
    public void setName(String name) { this.name = name; }
    public String getDescription() { return description; }
    public void setDescription(String description) { this.description = description; }
    public Long getCreatorId() { return creatorId; }
    public String getCreatorUsername() { return creatorUsername; }
    public int getMemberCount() { return memberCount; }
    public void setMemberCount(int memberCount) { this.memberCount = memberCount; }
    public LocalDateTime getCreatedAt() { return createdAt; }
}
