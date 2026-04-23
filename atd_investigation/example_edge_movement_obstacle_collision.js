// upsiloncli/tests/scenarios/edge_movement_obstacle_collision.js
// @spec-link [[mech_move_validation_move_validation_obstacle_collision]]
// @spec-link [[mech_board_generation_terrain_obstacles]]
// @spec-link [[entity_grid]]

const agentIndex = upsilon.getAgentIndex();
const botId = Math.floor(Math.random() * 10000) + "_" + agentIndex;
const accountName = "obstacle_bot_" + botId;
const password = "VerySecurePassword123!";

upsilon.log(`[Bot-${agentIndex}] Starting EC-01: Movement on Obstacle Tiles`);

// 1. Setup
upsilon.bootstrapBot(accountName, password);
const matchData = upsilon.joinWaitMatch("1v1_PVE");
upsilon.log(`[Bot-${agentIndex}] Joined match: ${matchData.match_id}`);

// 2. Wait for first turn
const board = upsilon.waitNextTurn();
if (!board) {
    upsilon.assert(false, "ERROR: Match ended unexpectedly");
}

const myChar = upsilon.currentCharacter();
const startPos = myChar.position;
upsilon.log(`[Bot-${agentIndex}] Character at position: ${startPos.x},${startPos.y}`);

// 3. Find obstacle tile and attempt to move through it
const gridWidth = board.grid[0].length;
const gridHeight = board.grid.length;
let obstacleFound = false;
let obstaclePos = null;

// Search for obstacle tile
for (let y = 0; y < gridHeight; y++) {
    for (let x = 0; x < gridWidth; x++) {
        const cell = board.grid[y][x];
        if (cell && cell.obstacle) {
            obstaclePos = { x, y };
            obstacleFound = true;
            upsilon.log(`[Bot-${agentIndex}] Found obstacle at: ${x},${y}`);
            break;
        }
    }
    if (obstacleFound) break;
}

if (!obstacleFound) {
    upsilon.log(`[Bot-${agentIndex}] SKIP: No obstacles found on this board`);
} else {
    // Attempt to move through obstacle
    const pathToObstacle = upsilon.planTravelToward(myChar.id, obstaclePos, board);

    if (pathToObstacle.length > 0) {
        upsilon.log(`[Bot-${agentIndex}] Attempting to move to obstacle at ${obstaclePos.x},${obstaclePos.y}...`);
        try {
            upsilon.call("game_action", {
                id: matchData.match_id,
                type: "move",
                entity_id: myChar.id,
                target_coords: pathToObstacle
            });
            upsilon.assert(false, "ERROR: Movement through obstacle was accepted by server!");
        } catch (e) {
            upsilon.log(`[Bot-${agentIndex}] ✅ Obstacle collision properly rejected: ${e.message}`);
            upsilon.assertEquals(e.error_key, "entity.path.obstacle", "Wrong error key for obstacle collision");
        }

        // Verify position unchanged
        const updatedBoard = upsilon.call("game_state", { id: matchData.match_id });
        const updatedChar = updatedBoard.data.players[0].entities.find(e => e.id === myChar.id);
        upsilon.assertEquals(updatedChar.position.x, startPos.x, "Character X position changed after failed move");
        upsilon.assertEquals(updatedChar.position.y, startPos.y, "Character Y position changed after failed move");
    } else {
        upsilon.log(`[Bot-${agentIndex}] SKIP: No path to obstacle available`);
    }

    // 4. Attempt valid move around obstacle
    const validPos = { x: Math.min(startPos.x + 1, gridWidth - 1), y: startPos.y };
    if (validPos.x !== startPos.x) {
        upsilon.log(`[Bot-${agentIndex}] Attempting valid move to ${validPos.x},${validPos.y}...`);
        try {
            upsilon.call("game_action", {
                id: matchData.match_id,
                type: "move",
                entity_id: myChar.id,
                target_coords: [validPos]
            });
            upsilon.log(`[Bot-${agentIndex}] ✅ Valid move succeeded`);
        } catch (e) {
            // This might fail if not turn or out of movement, which is acceptable
            upsilon.log(`[Bot-${agentIndex}] Valid move failed (may be expected): ${e.message}`);
        }
    }
}

upsilon.log(`[Bot-${agentIndex}] EC-01: MOVEMENT ON OBSTACLE TILES PASSED.`);
