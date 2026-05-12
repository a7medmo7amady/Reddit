package com.reddit.clone.model;
// User Model
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

import jakarta.persistence.CollectionTable;
import jakarta.persistence.Column;
import jakarta.persistence.ElementCollection;
import jakarta.persistence.Entity;
import jakarta.persistence.EnumType;
import jakarta.persistence.Enumerated;
import jakarta.persistence.FetchType;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.Table;

@Entity
@Table(name = "users")
public class User {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String username;

    @Column(nullable = false, unique = true)
    private String email;

    @Column(nullable = true)
    private String password;

    private String displayName;

    @Column(length = 500)
    private String bio;

    private String avatar;

    private String banner;

    @ElementCollection(fetch = FetchType.EAGER)
    @CollectionTable(name = "user_links", joinColumns = @JoinColumn(name = "user_id"))
    @Column(name = "link")
    private List<String> links = new ArrayList<>();

    private int postKarma = 0;
    private int commentKarma = 0;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private Role role;

    @Enumerated(EnumType.STRING)
    private OAuthProvider oauthProvider;

    @Column(nullable = false)
    private boolean isBanned = false;

    @Column(nullable = false, updatable = false)
    private LocalDateTime createdAt = LocalDateTime.now();

    protected User() {}

    public User(String username, String email, String password, String displayName, String bio,
                String avatar, String banner, Role role, OAuthProvider oauthProvider) {
        this.username = username;
        this.email = email;
        this.password = password;
        this.displayName = displayName;
        this.bio = bio;
        this.avatar = avatar;
        this.banner = banner;
        this.role = role;
        this.oauthProvider = oauthProvider;
    }

    public Long getId() { return id; }

    public String getUsername() { return username; }
    public void setUsername(String username) { this.username = username; }

    public String getEmail() { return email; }
    public void setEmail(String email) { this.email = email; }

    public String getPassword() { return password; }
    public void setPassword(String password) { this.password = password; }

    public String getDisplayName() { return displayName; }
    public void setDisplayName(String displayName) { this.displayName = displayName; }

    public String getBio() { return bio; }
    public void setBio(String bio) { this.bio = bio; }

    public String getAvatar() { return avatar; }
    public void setAvatar(String avatar) { this.avatar = avatar; }

    public String getBanner() { return banner; }
    public void setBanner(String banner) { this.banner = banner; }

    public List<String> getLinks() { return links; }
    public void setLinks(List<String> links) {
        if (links != null && links.size() > 3) {
            throw new IllegalArgumentException("A user can have at most 3 links");
        }
        this.links = links != null ? links : new ArrayList<>();
    }

    public int getPostKarma() { return postKarma; }
    public void setPostKarma(int postKarma) { this.postKarma = postKarma; }

    public int getCommentKarma() { return commentKarma; }
    public void setCommentKarma(int commentKarma) { this.commentKarma = commentKarma; }

    public Role getRole() { return role; }
    public void setRole(Role role) { this.role = role; }

    public OAuthProvider getOauthProvider() { return oauthProvider; }
    public void setOauthProvider(OAuthProvider oauthProvider) { this.oauthProvider = oauthProvider; }

    public boolean isBanned() { return isBanned; }
    public void setBanned(boolean isBanned) { this.isBanned = isBanned; }

    public LocalDateTime getCreatedAt() { return createdAt; }
}
