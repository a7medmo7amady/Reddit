package com.reddit.clone.dto;

import com.reddit.clone.model.OAuthProvider;
import com.reddit.clone.model.Role;

import java.time.LocalDateTime;
import java.util.List;

public record PrivateProfileResponse(
        Long id,
        String username,
        String email,
        String displayName,
        String bio,
        String avatar,
        String banner,
        List<String> links,
        int karma,
        Role role,
        OAuthProvider oauthProvider,
        LocalDateTime createdAt
) {}
