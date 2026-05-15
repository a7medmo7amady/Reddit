package com.reddit.clone.model;

import jakarta.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "user_communities", uniqueConstraints = {
    @UniqueConstraint(columnNames = {"user_id", "community_id"})
})
public class UserCommunity {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "community_id", nullable = false)
    private Community community;

    @Column(nullable = false, updatable = false)
    private LocalDateTime joinedAt = LocalDateTime.now();

    protected UserCommunity() {}

    public UserCommunity(User user, Community community) {
        this.user = user;
        this.community = community;
    }

    public Long getId() { return id; }
    public User getUser() { return user; }
    public Community getCommunity() { return community; }
    public LocalDateTime getJoinedAt() { return joinedAt; }
}
