package com.reddit.clone.model;

import jakarta.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "community_bans",
       uniqueConstraints = @UniqueConstraint(columnNames = {"user_id", "community_name"}))
public class CommunityBan {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "user_id", nullable = false)
    private Long userId;

    @Column(name = "community_name", nullable = false)
    private String communityName;

    @Column(name = "banned_by", nullable = false)
    private Long bannedBy;

    @Column(length = 500)
    private String reason;

    @Column(nullable = false, updatable = false)
    private LocalDateTime bannedAt = LocalDateTime.now();

    protected CommunityBan() {}

    public CommunityBan(Long userId, String communityName, Long bannedBy, String reason) {
        this.userId        = userId;
        this.communityName = communityName;
        this.bannedBy      = bannedBy;
        this.reason        = reason;
    }

    public Long getId()                  { return id; }
    public Long getUserId()              { return userId; }
    public String getCommunityName()     { return communityName; }
    public Long getBannedBy()            { return bannedBy; }
    public String getReason()            { return reason; }
    public LocalDateTime getBannedAt()   { return bannedAt; }
}
