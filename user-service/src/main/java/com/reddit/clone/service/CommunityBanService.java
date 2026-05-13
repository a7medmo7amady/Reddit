package com.reddit.clone.service;

import com.reddit.clone.event.CommunityBanEvent;
import com.reddit.clone.exception.UserNotFoundException;
import com.reddit.clone.kafka.BanEventPublisher;
import com.reddit.clone.model.CommunityBan;
import com.reddit.clone.model.User;
import com.reddit.clone.repository.CommunityBanRepository;
import com.reddit.clone.repository.UserRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;

@Service
public class CommunityBanService {

    private final CommunityBanRepository banRepository;
    private final UserRepository         userRepository;
    private final BanEventPublisher      publisher;

    public CommunityBanService(CommunityBanRepository banRepository,
                               UserRepository userRepository,
                               BanEventPublisher publisher) {
        this.banRepository = banRepository;
        this.userRepository = userRepository;
        this.publisher      = publisher;
    }

    @Transactional
    public void ban(String targetUsername, String communityName, Long moderatorId, String reason) {
        User target = userRepository.findByUsername(targetUsername)
                .orElseThrow(() -> new UserNotFoundException("User not found: " + targetUsername));

        if (banRepository.existsByUserIdAndCommunityName(target.getId(), communityName)) {
            return; 
        }

        banRepository.save(new CommunityBan(target.getId(), communityName, moderatorId, reason));

        publisher.publish(new CommunityBanEvent(
                target.getId(),
                target.getUsername(),
                communityName,
                "BANNED",
                reason,
                LocalDateTime.now().toString()
        ));
    }

    @Transactional
    public void unban(String targetUsername, String communityName) {
        User target = userRepository.findByUsername(targetUsername)
                .orElseThrow(() -> new UserNotFoundException("User not found: " + targetUsername));

        banRepository.findByUserIdAndCommunityName(target.getId(), communityName)
                .ifPresent(ban -> {
                    banRepository.delete(ban);
                    publisher.publish(new CommunityBanEvent(
                            target.getId(),
                            target.getUsername(),
                            communityName,
                            "UNBANNED",
                            null,
                            LocalDateTime.now().toString()
                    ));
                });
    }

    public boolean isBanned(Long userId, String communityName) {
        return banRepository.existsByUserIdAndCommunityName(userId, communityName);
    }
}
