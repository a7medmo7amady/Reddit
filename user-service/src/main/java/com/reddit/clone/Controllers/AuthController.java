package com.reddit.clone.Controllers;

import com.reddit.clone.dto.AuthResponse;
import com.reddit.clone.dto.LoginRequest;
import com.reddit.clone.dto.SignupRequest;
import com.reddit.clone.exception.InvalidCredentialsException;
import com.reddit.clone.model.User;
import com.reddit.clone.security.JWTservice;
import com.reddit.clone.security.TokenService;
import com.reddit.clone.service.AuthService;
import jakarta.servlet.http.HttpServletResponse;
import jakarta.validation.Valid;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseCookie;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.Duration;

@RestController
@RequestMapping("/auth")
public class AuthController {

    private final AuthService  authService;
    private final JWTservice   jwtService;
    private final TokenService tokenService;

    public AuthController(AuthService authService, JWTservice jwtService, TokenService tokenService) {
        this.authService  = authService;
        this.jwtService   = jwtService;
        this.tokenService = tokenService;
    }

    @PostMapping("/signup")
    public ResponseEntity<AuthResponse> signup(@Valid @RequestBody SignupRequest req,
                                               HttpServletResponse response) {
        User user = authService.signup(req.username(), req.email(), req.password());
        return issue(user, response, HttpStatus.CREATED);
    }

    @PostMapping("/login")
    public ResponseEntity<AuthResponse> login(@Valid @RequestBody LoginRequest req,
                                              HttpServletResponse response) {
        User user = authService.login(req.email(), req.password());
        return issue(user, response, HttpStatus.OK);
    }

    @PostMapping("/refresh")
    public ResponseEntity<AuthResponse> refresh(
            @CookieValue(name = "refresh_token", required = false) String oldToken,
            HttpServletResponse response) {
        if (oldToken == null) throw new InvalidCredentialsException("No refresh token provided");
        TokenService.RotationResult rotation = tokenService.rotate(oldToken);
        User user = authService.findById(rotation.userId());
        String accessToken = jwtService.generateAccessToken(user);
        setRefreshCookie(response, rotation.newToken());
        return ResponseEntity.ok(new AuthResponse(accessToken));
    }

    @PostMapping("/logout")
    public ResponseEntity<Void> logout(
            @CookieValue(name = "refresh_token", required = false) String token,
            HttpServletResponse response) {
        if (token != null) tokenService.revoke(token);
        clearRefreshCookie(response);
        return ResponseEntity.noContent().build();
    }

    // ── helpers ──────────────────────────────────────────────────────────────

    private ResponseEntity<AuthResponse> issue(User user, HttpServletResponse response, HttpStatus status) {
        String refreshToken = tokenService.create(user.getId());
        String accessToken  = jwtService.generateAccessToken(user);
        setRefreshCookie(response, refreshToken);
        return ResponseEntity.status(status).body(new AuthResponse(accessToken));
    }

    private void setRefreshCookie(HttpServletResponse response, String token) {
        response.addHeader(HttpHeaders.SET_COOKIE, buildCookie(token, Duration.ofDays(7)));
    }

    private void clearRefreshCookie(HttpServletResponse response) {
        response.addHeader(HttpHeaders.SET_COOKIE, buildCookie("", Duration.ZERO));
    }

    private String buildCookie(String value, Duration maxAge) {
        return ResponseCookie.from("refresh_token", value)
                .httpOnly(true)
                .secure(false)   // set true behind HTTPS in production
                .path("/auth")
                .maxAge(maxAge)
                .sameSite("Strict")
                .build()
                .toString();
    }
}
