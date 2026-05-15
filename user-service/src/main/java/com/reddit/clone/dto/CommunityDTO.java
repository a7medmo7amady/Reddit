package com.reddit.clone.dto;

import java.io.Serializable;

public record CommunityDTO(Long id, String name, String description, int memberCount) implements Serializable {}
