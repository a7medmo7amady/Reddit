package com.reddit.clone.model;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.index.Indexed;
import org.springframework.data.mongodb.core.mapping.Document;

import java.util.ArrayList;
import java.util.List;

@Document(collection = "user_activities")
public class UserActivity {

    @Id
    private String id;

    @Indexed(unique = true)
    private Long userId;

    private List<String> savedPostIds = new ArrayList<>();
    private List<String> hiddenPostIds = new ArrayList<>();
    private List<String> upvotedPostIds = new ArrayList<>();

    protected UserActivity() {}

    public UserActivity(Long userId) {
        this.userId = userId;
    }

    public String getId() { return id; }

    public Long getUserId() { return userId; }

    public List<String> getSavedPostIds() { return savedPostIds; }
    public void setSavedPostIds(List<String> savedPostIds) { this.savedPostIds = savedPostIds; }

    public List<String> getHiddenPostIds() { return hiddenPostIds; }
    public void setHiddenPostIds(List<String> hiddenPostIds) { this.hiddenPostIds = hiddenPostIds; }

    public List<String> getUpvotedPostIds() { return upvotedPostIds; }
    public void setUpvotedPostIds(List<String> upvotedPostIds) { this.upvotedPostIds = upvotedPostIds; }
}
