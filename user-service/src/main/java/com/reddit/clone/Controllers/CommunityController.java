package com.reddit.clone.Controllers;

import com.reddit.clone.dto.CommunityDTO;
import com.reddit.clone.dto.CreateCommunityRequest;
import com.reddit.clone.model.User;
import com.reddit.clone.service.CommunityService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/communities")
public class CommunityController {

    private final CommunityService communityService;

    public CommunityController(CommunityService communityService) {
        this.communityService = communityService;
    }

    @PostMapping
    public ResponseEntity<CommunityDTO> create(@Valid @RequestBody CreateCommunityRequest req,
                                               Authentication auth) {
        User user = (User) auth.getPrincipal();
        CommunityDTO created = communityService.createCommunity(user.getId(), req);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    @GetMapping("/{name}")
    public ResponseEntity<CommunityDTO> getCommunity(@PathVariable String name) {
        return ResponseEntity.ok(communityService.getCommunity(name));
    }

    @GetMapping("/me")
    public ResponseEntity<List<CommunityDTO>> getMyCommunitites(Authentication auth) {
        User user = (User) auth.getPrincipal();
        return ResponseEntity.ok(communityService.getFollowedCommunities(user.getId()));
    }

    @PostMapping("/{name}/join")
    public ResponseEntity<Void> join(@PathVariable String name, Authentication auth) {
        User user = (User) auth.getPrincipal();
        communityService.joinCommunity(user.getId(), name);
        return ResponseEntity.ok().build();
    }

    @PostMapping("/{name}/leave")
    public ResponseEntity<Void> leave(@PathVariable String name, Authentication auth) {
        User user = (User) auth.getPrincipal();
        communityService.leaveCommunity(user.getId(), name);
        return ResponseEntity.ok().build();
    }

    @GetMapping("/{name}/membership")
    public ResponseEntity<Map<String, Boolean>> membership(@PathVariable String name, Authentication auth) {
        User user = (User) auth.getPrincipal();
        boolean member = communityService.isMember(user.getId(), name);
        return ResponseEntity.ok(Map.of("member", member));
    }
}
