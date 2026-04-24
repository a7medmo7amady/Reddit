package com.reddit.clone.model;

import jakarta.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "refresh_tokens")
public class RefreshToken {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String tokenHash;

    @Column(nullable = false)
    private Long userId;

    @Column(nullable = false)
    private LocalDateTime expiresAt;

    @Column(nullable = false)
    private boolean revoked = false;

    @Column(nullable = false, updatable = false)
    private LocalDateTime createdAt = LocalDateTime.now();

    protected RefreshToken() {}

    public RefreshToken(String tokenHash, Long userId, LocalDateTime expiresAt) {
        this.tokenHash = tokenHash;
        this.userId    = userId;
        this.expiresAt = expiresAt;
    }

    public Long getId()              { return id; }
    public String getTokenHash()     { return tokenHash; }
    public Long getUserId()          { return userId; }
    public LocalDateTime getExpiresAt() { return expiresAt; }
    public boolean isRevoked()       { return revoked; }
    public void setRevoked(boolean revoked) { this.revoked = revoked; }
    public LocalDateTime getCreatedAt() { return createdAt; }
}
