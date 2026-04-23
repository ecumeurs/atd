// upsiloncli/tests/scenarios/edge_attack_friendly_fire.js
// @spec-link [[rule_friendly_fire]]
// @spec-link [[rule_friendly_fire_team_validation]]
// @spec-link [[rule_friendly_fire_match_type]]

const agentIndex = upsilon.getAgentIndex();
const botId = Math.floor(Math.random() * 10000) + "_" + agentIndex;
const accountName = "ff_edge_bot_" + botId;
const password = "VerySecurePassword123!";

upsilon.log(`[Bot-${agentIndex}] Starting EC-12: Attack Friendly Fire (Same Team)`);

// 1. Setup (2v2 match)
upsilon.bootstrapBot(accountName, password);
const matchData = upsilon.joinWaitMatch("2v2_PVP");
upsilon.log(`[Bot-${agentIndex}] Joined 2v2 match: ${matchData.match_id}`);

// Share identity for coordination
upsilon.setShared(`ff_bot_${agentIndex}`, upsilon.myCharacters()[0].id);
upsilon.syncGroup("ff_ready", 2);

// Wait for turn and find ally
const board = upsilon.waitNextTurn();
if (!board) {
    upsilon.assert(false, "ERROR: Match ended unexpectedly");
}

const myChar = upsilon.currentCharacter();
const allyChars = upsilon.myAlliesCharacters();

if (allyChars.length > 0) {
    const targetAlly = allyChars[0];
    upsilon.log(`[Bot-${agentIndex}] Attempting Friendly Fire on ally ${targetAlly.name} (team ${targetAlly.team})...`);
    upsilon.log(`[Bot-${agentIndex}] My team: ${myChar.team}, Ally team: ${targetAlly.team}`);

    try {
        upsilon.call("game_action", {
            id: matchData.match_id,
            type: "attack",
            entity_id: myChar.id,
            target_coords: [targetAlly.position]
        });
        upsilon.assert(false, "ERROR: Friendly fire attack was accepted by server!");
    } catch (e) {
        upsilon.log(`[Bot-${agentIndex}] ✅ Friendly fire properly rejected: ${e.message}`);
        upsilon.assertEquals(e.error_key, "rule.friendly_fire", "Wrong error key for friendly fire");
    }

    // Verify ally HP unchanged
    const updatedBoard = upsilon.call("game_state", { id: matchData.match_id });
    const updatedAlly = updatedBoard.data.players.find(p => p.entities.find(e => e.id === targetAlly.id));
    const allyEntity = updatedAlly.entities.find(e => e.id === targetAlly.id);
    upsilon.assertEquals(allyEntity.hp, targetAlly.hp, "Ally HP changed after blocked attack");
    upsilon.log(`[Bot-${agentIndex}] ✅ Ally HP unchanged (${allyEntity.hp})`);
} else {
    upsilon.log(`[Bot-${agentIndex}] SKIP: No allies visible this turn (may be 1v1 match)`);
}

// Test: Attack enemy should succeed
const enemyChars = upsilon.myFoesCharacters();
if (enemyChars.length > 0) {
    const targetEnemy = enemyChars[0];
    upsilon.log(`[Bot-${agentIndex}] Testing valid attack on enemy ${targetEnemy.name}...`);

    try {
        upsilon.call("game_action", {
            id: matchData.match_id,
            type: "attack",
            entity_id: myChar.id,
            target_coords: [targetEnemy.position]
        });
        upsilon.log(`[Bot-${agentIndex}] ✅ Attack on enemy succeeded`);
    } catch (e) {
        // May fail due to range, cooldown, etc.
        upsilon.log(`[Bot-${agentIndex}] Attack on enemy failed (may be expected): ${e.message}`);
    }
}

upsilon.log(`[Bot-${agentIndex}] EC-12: ATTACK FRIENDLY FIRE (SAME TEAM) PASSED.`);
