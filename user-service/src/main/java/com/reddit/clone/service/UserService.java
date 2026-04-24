package com.reddit.clone.service;

import com.reddit.clone.dto.PrivateProfileResponse;
import com.reddit.clone.dto.PublicProfileResponse;
import com.reddit.clone.dto.UpdateProfileRequest;
import com.reddit.clone.exception.UserNotFoundException;
import com.reddit.clone.model.User;
import com.reddit.clone.repository.UserRepository;
import com.reddit.clone.security.TokenService;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class UserService {

    private final UserRepository userRepository;
    private final TokenService   tokenService;

    public UserService(UserRepository userRepository, TokenService tokenService) {
        this.userRepository = userRepository;
        this.tokenService   = tokenService;
    }

    public User findById(Long id) {
        return userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException("User not found with id: " + id));
    }

    public User findByUsername(String username) {
        return userRepository.findByUsername(username)
                .orElseThrow(() -> new UserNotFoundException("User not found: " + username));
    }

    public PublicProfileResponse getPublicProfile(String username) {
        User u = findByUsername(username);
        return new PublicProfileResponse(
                u.getUsername(), u.getDisplayName(), u.getBio(),
                u.getAvatar(), u.getBanner(),
                u.getPostKarma() + u.getCommentKarma(),
                u.getCreatedAt());
    }

    public PrivateProfileResponse toPrivateProfile(User u) {
        return new PrivateProfileResponse(
                u.getId(), u.getUsername(), u.getEmail(),
                u.getDisplayName(), u.getBio(), u.getAvatar(), u.getBanner(),
                u.getLinks(), u.getPostKarma() + u.getCommentKarma(),
                u.getRole(), u.getOauthProvider(), u.getCreatedAt());
    }

    @Transactional
    public User updateProfile(Long userId, UpdateProfileRequest req) {
        User user = findById(userId);
        if (req.displayName() != null) user.setDisplayName(req.displayName());
        if (req.bio()         != null) user.setBio(req.bio());
        if (req.avatar()      != null) user.setAvatar(req.avatar());
        if (req.banner()      != null) user.setBanner(req.banner());
        if (req.links()       != null) user.setLinks(req.links());
        return userRepository.save(user);
    }

    @Transactional
    public void deleteAccount(Long userId) {
        if (!userRepository.existsById(userId)) {
            throw new UserNotFoundException("User not found with id: " + userId);
        }
        tokenService.revokeAllForUser(userId);
        userRepository.deleteById(userId);
    }
}
