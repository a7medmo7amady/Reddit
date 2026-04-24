package com.reddit.clone.dto;

import java.time.LocalDateTime;

public record PublicProfileResponse(
        String username,
        String displayName,
        String bio,
        String avatar,
        String banner,
        int karma,
        LocalDateTime createdAt
) {}
