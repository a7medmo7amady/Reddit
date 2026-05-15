package com.reddit.clone.Controllers;

import com.reddit.clone.model.User;
import com.reddit.clone.service.CommunityBanService;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.Authentication;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/communities/{community}/bans")
public class CommunityBanController {

    private final CommunityBanService banService;

    public CommunityBanController(CommunityBanService banService) {
        this.banService = banService;
    }

    @PostMapping("/{username}")
    public ResponseEntity<Void> ban(
            @PathVariable String community,
            @PathVariable String username,
            @RequestParam(required = false) String reason,
            Authentication auth) {
        User mod = (User) auth.getPrincipal();
        banService.ban(username, community, mod.getId(), reason);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/{username}")
    public ResponseEntity<Void> unban(
            @PathVariable String community,
            @PathVariable String username) {
        banService.unban(username, community);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/{username}")
    public ResponseEntity<Boolean> isBanned(
            @PathVariable String community,
            @PathVariable String username,
            Authentication auth) {
        User requester = (User) auth.getPrincipal();
        boolean banned = banService.isBanned(requester.getId(), community);
        return ResponseEntity.ok(banned);
    }
}
