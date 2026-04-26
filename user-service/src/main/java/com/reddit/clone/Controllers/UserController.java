package com.reddit.clone.Controllers;

import com.reddit.clone.dto.PrivateProfileResponse;
import com.reddit.clone.dto.PublicProfileResponse;
import com.reddit.clone.dto.UpdateProfileRequest;
import com.reddit.clone.model.User;
import com.reddit.clone.service.UserService;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/users")
public class UserController {

    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    @GetMapping("/{username}")
    public ResponseEntity<PublicProfileResponse> getPublicProfile(@PathVariable String username) {
        return ResponseEntity.ok(userService.getPublicProfile(username));
    }

    @GetMapping("/me")
    public ResponseEntity<PrivateProfileResponse> getMyProfile(Authentication auth) {
        User user = (User) auth.getPrincipal();
        return ResponseEntity.ok(userService.toPrivateProfile(user));
    }

    @PatchMapping("/me")
    public ResponseEntity<PrivateProfileResponse> updateProfile(
            @Valid @RequestBody UpdateProfileRequest req,
            Authentication auth) {
        User user    = (User) auth.getPrincipal();
        User updated = userService.updateProfile(user.getId(), req);
        return ResponseEntity.ok(userService.toPrivateProfile(updated));
    }

    @PostMapping("/block/{username}")
    public ResponseEntity<Void> blockUser(@PathVariable String username, Authentication auth) {
        User user = (User) auth.getPrincipal();
        userService.blockUser(user.getId(), username);
        return ResponseEntity.ok().build();
    }

    @GetMapping("/me/blocked")
    public ResponseEntity<List<PublicProfileResponse>> getBlockedUsers(Authentication auth) {
        User user = (User) auth.getPrincipal();
        return ResponseEntity.ok(userService.getBlockedUsers(user.getId()));
    }

    @PostMapping("/follow/{username}")
    public ResponseEntity<Void> followUser(@PathVariable String username, Authentication auth) {
        User user = (User) auth.getPrincipal();
        userService.followUser(user.getId(), username);
        return ResponseEntity.ok().build();
    }

    @PostMapping("/unblock/{username}")
    public ResponseEntity<Void> unblockUser(@PathVariable String username, Authentication auth) {
        User user = (User) auth.getPrincipal();
        userService.unblockUser(user.getId(), username);
        return ResponseEntity.ok().build();
    }

    @PostMapping("/unfollow/{username}")
    public ResponseEntity<Void> unfollowUser(@PathVariable String username, Authentication auth) {
        User user = (User) auth.getPrincipal();
        userService.unfollowUser(user.getId(), username);
        return ResponseEntity.ok().build();
    }

    @GetMapping("/me/following")
    public ResponseEntity<List<PublicProfileResponse>> getFollowing(Authentication auth) {
        User user = (User) auth.getPrincipal();
        return ResponseEntity.ok(userService.getFollowing(user.getId()));
    }

    @GetMapping("/me/followers")
    public ResponseEntity<List<PublicProfileResponse>> getFollowers(Authentication auth) {
        User user = (User) auth.getPrincipal();
        return ResponseEntity.ok(userService.getFollowers(user.getId()));
    }

    @DeleteMapping("/me")
    public ResponseEntity<Void> deleteAccount(Authentication auth) {
        User user = (User) auth.getPrincipal();
        userService.deleteAccount(user.getId());
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/admin/ban/{username}")
    public ResponseEntity<Void> banUser(@PathVariable String username, Authentication auth) {
        User admin = (User) auth.getPrincipal();
        userService.banUser(admin.getId(), username);
        return ResponseEntity.ok().build();
    }

    @PostMapping("/admin/unban/{username}")
    public ResponseEntity<Void> unbanUser(@PathVariable String username, Authentication auth) {
        User admin = (User) auth.getPrincipal();
        userService.unbanUser(admin.getId(), username);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/admin/remove/{username}")
    public ResponseEntity<Void> removeUserAsAdmin(@PathVariable String username, Authentication auth) {
        User admin = (User) auth.getPrincipal();
        userService.removeUser(admin.getId(), username);
        return ResponseEntity.noContent().build();
    }
}
