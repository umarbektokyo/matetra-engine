# matetra
The ultimate math game made by self-claimed "The best future mathematicians on this tiny planet."
```
                         $$\                $$\                        
                         $$ |               $$ |                       
$$$$$$\$$$$\   $$$$$$\ $$$$$$\    $$$$$$\ $$$$$$\    $$$$$$\  $$$$$$\  
$$  _$$  _$$\  \____$$\\_$$  _|  $$  __$$\\_$$  _|  $$  __$$\ \____$$\ 
$$ / $$ / $$ | $$$$$$$ | $$ |    $$$$$$$$ | $$ |    $$ |  \__|$$$$$$$ |
$$ | $$ | $$ |$$  __$$ | $$ |$$\ $$   ____| $$ |$$\ $$ |     $$  __$$ |
$$ | $$ | $$ |\$$$$$$$ | \$$$$  |\$$$$$$$\  \$$$$  |$$ |     \$$$$$$$ |
\__| \__| \__| \_______|  \____/  \_______|  \____/ \__|      \_______|

```

# Installation
```bash
# For the client
go install github.com/umarbektokyo/matetra-engine/cmd/matetra-client@latest
# For the server
go install github.com/umarbektokyo/matetra-engine/cmd/matetra-server@latest
```
> WARNING: Ensure you have `$GOPATH/bin` in `PATH`
# Running
```bash
# For the client
matetra-client <server-address>:1729
# ex: matetra-client localhost:1729

# For the server
matetra-server start <game-title>
# ex: matetra-server start WonderfulGame
```

# Rules & Procedures
- Don’t cheat.
- At the beginning of the game and after every turn, players will have 6 cards.
- A player at any point in time will either be attacking or defending.
- On their turn (defending), the player will be able to add numbers to their set, and defend those numbers. - On other’s turn (attacker), the player can attack the numbers in an effort to either remove the playing user’s number or decrease it.
- Defending entails that a player can add numbers to their set and use theorems and functions on the numbers already in their set.
- Attacking entails that a player can NOT add numbers to their own set and they may only OWN use theorems, functions and numbers on the numbers in the set of the player currently defending.
- Player who has the highest number at the end of the game wins!

Variable Meanings:
d → number must be rolled on a dice
a → the “main number”, if attacking, this is where the attacked number goes.
b,c,… → You can only use your own numbers from the set, they will be deleted once used.
u → call umarbek and ask to generate a random number.]

# Welcome the crew!
- Flush! - Esia
- Noga L.
- Pikachu - Meharwan
- Umarbek
