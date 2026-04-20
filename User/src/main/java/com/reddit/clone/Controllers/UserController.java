package com.reddit.clone.Controllers;

import com.reddit.clone.model.User;
import com.reddit.clone.service.UserService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("/auth")
public class UserController {

    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    @PostMapping("/register")
    public ResponseEntity<?> register(@RequestBody Map<String, String> body) {
        String username = body.get("username");
        String email    = body.get("email");
        String password = body.get("password");

        User user = userService.register(username, email, password);

        return ResponseEntity.status(HttpStatus.CREATED).body(Map.of(
            "id",       user.getId(),
            "username", user.getUsername(),
            "email",    user.getEmail()
        ));
    }
}
