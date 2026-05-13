package com.reddit.clone.service;

import com.reddit.clone.dto.PublicProfileResponse;
import com.reddit.clone.model.Community;
import com.reddit.clone.repository.CommunityRepository;
import com.reddit.clone.repository.UserRepository;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

@Service
public class SearchService {

    private final UserRepository userRepository;
    private final CommunityRepository communityRepository;
    private final UserService userService;

    public SearchService(UserRepository userRepository, 
                         CommunityRepository communityRepository,
                         UserService userService) {
        this.userRepository = userRepository;
        this.communityRepository = communityRepository;
        this.userService = userService;
    }

    public List<PublicProfileResponse> searchUsers(String query) {
        return userRepository.findByUsernameContainingIgnoreCase(query)
                .stream()
                .map(userService::toPublicProfile)
                .collect(Collectors.toList());
    }

    public List<Community> searchCommunities(String query) {
        return communityRepository.searchCommunities(query);
    }
}
