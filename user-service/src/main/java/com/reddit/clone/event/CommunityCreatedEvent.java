package com.reddit.clone.event;

import com.fasterxml.jackson.annotation.JsonProperty;

public record CommunityCreatedEvent(
        @JsonProperty("id")               String id,
        @JsonProperty("postgres_id")      Long postgresId,
        @JsonProperty("name")             String name,
        @JsonProperty("description")      String description,
        @JsonProperty("member_count")     int memberCount,
        @JsonProperty("creator_username") String creatorUsername,
        @JsonProperty("created_at")       String createdAt
) {}
