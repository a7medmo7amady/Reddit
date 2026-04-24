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

    @DeleteMapping("/me")
    public ResponseEntity<Void> deleteAccount(Authentication auth) {
        User user = (User) auth.getPrincipal();
        userService.deleteAccount(user.getId());
        return ResponseEntity.noContent().build();
    }
}
