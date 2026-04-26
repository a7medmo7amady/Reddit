package com.reddit.clone.service;

import com.reddit.clone.dto.PrivateProfileResponse;
import com.reddit.clone.dto.PublicProfileResponse;
import com.reddit.clone.dto.UpdateProfileRequest;
import com.reddit.clone.exception.UserNotFoundException;
import com.reddit.clone.model.User;
import com.reddit.clone.repository.UserRepository;
import com.reddit.clone.security.TokenService;
import com.reddit.clone.model.UserBlock;
import com.reddit.clone.repository.UserBlockRepository;
import com.reddit.clone.model.UserFollow;
import com.reddit.clone.repository.UserFollowRepository;
import com.reddit.clone.model.Role;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class UserService {

    private final UserRepository userRepository;
    private final TokenService   tokenService;
    private final UserBlockRepository userBlockRepository;
    private final UserFollowRepository userFollowRepository;

    public UserService(UserRepository userRepository, TokenService tokenService, UserBlockRepository userBlockRepository, UserFollowRepository userFollowRepository) {
        this.userRepository = userRepository;
        this.tokenService   = tokenService;
        this.userBlockRepository = userBlockRepository;
        this.userFollowRepository = userFollowRepository;
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
    public void blockUser(Long blockerId, String blockedUsername) {
        User blocker = findById(blockerId);
        User blocked = findByUsername(blockedUsername);

        if (blocker.getId().equals(blocked.getId())) {
            throw new IllegalArgumentException("You cannot block yourself");
        }

        if (!userBlockRepository.existsByBlockerAndBlocked(blocker, blocked)) {
            UserBlock userBlock = new UserBlock(blocker, blocked);
            userBlockRepository.save(userBlock);
        }
    }

    @Transactional
    public void unblockUser(Long blockerId, String blockedUsername) {
        User blocker = findById(blockerId);
        User blocked = findByUsername(blockedUsername);

        userBlockRepository.findByBlockerAndBlocked(blocker, blocked)
                .ifPresent(userBlockRepository::delete);
    }

    @Transactional(readOnly = true)
    public List<PublicProfileResponse> getBlockedUsers(Long blockerId) {
        User blocker = findById(blockerId);
        return userBlockRepository.findByBlocker(blocker).stream()
                .map(UserBlock::getBlocked)
                .map(u -> new PublicProfileResponse(
                        u.getUsername(), u.getDisplayName(), u.getBio(),
                        u.getAvatar(), u.getBanner(),
                        u.getPostKarma() + u.getCommentKarma(),
                        u.getCreatedAt()))
                .toList();
    }

    @Transactional
    public void followUser(Long followerId, String followedUsername) {
        User follower = findById(followerId);
        User followed = findByUsername(followedUsername);

        if (follower.getId().equals(followed.getId())) {
            throw new IllegalArgumentException("You cannot follow yourself");
        }

        if (!userFollowRepository.existsByFollowerAndFollowed(follower, followed)) {
            UserFollow userFollow = new UserFollow(follower, followed);
            userFollowRepository.save(userFollow);
        }
    }

    @Transactional
    public void unfollowUser(Long followerId, String followedUsername) {
        User follower = findById(followerId);
        User followed = findByUsername(followedUsername);

        userFollowRepository.findByFollowerAndFollowed(follower, followed)
                .ifPresent(userFollowRepository::delete);
    }

    @Transactional(readOnly = true)
    public List<PublicProfileResponse> getFollowing(Long followerId) {
        User follower = findById(followerId);
        return userFollowRepository.findByFollower(follower).stream()
                .map(UserFollow::getFollowed)
                .map(u -> new PublicProfileResponse(
                        u.getUsername(), u.getDisplayName(), u.getBio(),
                        u.getAvatar(), u.getBanner(),
                        u.getPostKarma() + u.getCommentKarma(),
                        u.getCreatedAt()))
                .toList();
    }

    @Transactional(readOnly = true)
    public List<PublicProfileResponse> getFollowers(Long userId) {
        User user = findById(userId);
        return userFollowRepository.findByFollowed(user).stream()
                .map(UserFollow::getFollower)
                .map(u -> new PublicProfileResponse(
                        u.getUsername(), u.getDisplayName(), u.getBio(),
                        u.getAvatar(), u.getBanner(),
                        u.getPostKarma() + u.getCommentKarma(),
                        u.getCreatedAt()))
                .toList();
    }

    @Transactional
    public void banUser(Long adminId, String targetUsername) {
        User admin = findById(adminId);
        if (admin.getRole() != Role.ADMIN) {
            throw new SecurityException("Only admins can ban users");
        }
        User target = findByUsername(targetUsername);
        target.setBanned(true);
        userRepository.save(target);
    }

    @Transactional
    public void unbanUser(Long adminId, String targetUsername) {
        User admin = findById(adminId);
        if (admin.getRole() != Role.ADMIN) {
            throw new SecurityException("Only admins can unban users");
        }
        User target = findByUsername(targetUsername);
        target.setBanned(false);
        userRepository.save(target);
    }

    @Transactional
    public void removeUser(Long adminId, String targetUsername) {
        User admin = findById(adminId);
        if (admin.getRole() != Role.ADMIN) {
            throw new SecurityException("Only admins can remove users");
        }
        User target = findByUsername(targetUsername);
        tokenService.revokeAllForUser(target.getId());
        userRepository.deleteById(target.getId());
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
