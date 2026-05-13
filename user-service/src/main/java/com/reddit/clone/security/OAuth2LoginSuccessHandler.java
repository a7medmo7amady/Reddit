package com.reddit.clone.security;

import com.reddit.clone.model.OAuthProvider;
import com.reddit.clone.model.User;
import com.reddit.clone.service.AuthService;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseCookie;
import org.springframework.security.core.Authentication;
import org.springframework.security.oauth2.core.user.OAuth2User;
import org.springframework.security.web.authentication.AuthenticationSuccessHandler;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.time.Duration;

@Component
public class OAuth2LoginSuccessHandler implements AuthenticationSuccessHandler {

    private final AuthService authService;
    private final JWTservice jwtService;
    private final TokenService tokenService;

    @Value("${client.redirect-url:http://localhost:3000}")
    private String clientRedirectUrl;

    public OAuth2LoginSuccessHandler(AuthService authService,
                                     JWTservice jwtService,
                                     TokenService tokenService) {
        this.authService = authService;
        this.jwtService = jwtService;
        this.tokenService = tokenService;
    }

    @Override
    public void onAuthenticationSuccess(HttpServletRequest request,
                                        HttpServletResponse response,
                                        Authentication authentication) throws IOException, ServletException {
        OAuth2User oauthUser = (OAuth2User) authentication.getPrincipal();
        String email = oauthUser.getAttribute("email");
        String name = oauthUser.getAttribute("name");
        String picture = oauthUser.getAttribute("picture");

        User user = authService.findOrCreateOAuthUser(email, name, picture, OAuthProvider.GOOGLE);
        String refreshToken = tokenService.create(user.getId());
        String accessToken = jwtService.generateAccessToken(user);

        response.addHeader(HttpHeaders.SET_COOKIE, refreshCookie(refreshToken));
        response.sendRedirect(clientRedirectUrl + "?accessToken=" + accessToken);
    }

    private String refreshCookie(String token) {
        return ResponseCookie.from("refresh_token", token)
                .httpOnly(true)
                .secure(false)
                .path("/auth")
                .maxAge(Duration.ofDays(7))
                .sameSite("Strict")
                .build()
                .toString();
    }
}
