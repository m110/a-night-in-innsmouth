# A Night in Innsmouth

This game has been created for the [GitHub Game Off 2024](https://itch.io/jam/game-off-2024).

It's a narrative game based on the prose of H. P. Lovecraft.

All graphical assets were created by [Małgorzata Bocian](https://www.behance.net/gosiabocianart).
The game engine was created by Miłosz Smółka (me) in Go using [Ebitengine](https://github.com/hajimehoshi/ebiten) and [donburi](https://github.com/yohamta/donburi).

State: an early prototype. The gameplay is mostly there, but the story is quite minimal, with many dialogs missing. To be added. :)

## License

The game code is licensed under the MIT License.
The assets in the `assets/game` directory are protected by copyright law.

In short, feel free to use the engine to create your own game, but bring your own assets and story.

## Editor

The game files are made of two kinds of files: story and levels.

The story is kept in the `twee` format exported from [Twine](https://twinery.org/).
It keeps all dialogs and choices the player can make.

Levels can be edited by [Tiled](https://mapeditor.org).
They keep the position of POIs on the levels, etc.

### Story

The format is loosely based on the `chapbook` format,
but it's extended with more features to make it more fit for the game.

The story is made of *passages*, some of which are connected.
A single passage can contain:

* conditions — the passage is not available until they're all met.
* macros — actions that happen once the passage is visited.
* paragraphs — text displayed to the player.
* links — dialog options the player chooses to continue.

Conditions and macros come first and are separated by two minuses (`--`).

Example:

```
if: fact arrivedHome
addItem: Favorite Stopwatch
--
You're finally home. You can relax now.

> [[Go to bed->Bed]]
> [[Go to the kitchen]]
```

#### Conditions

Conditions start with the `if:` or `unless:` prefix.

| Condition | Description                                               |
|-----------|-----------------------------------------------------------|
| fact      | Checks if the given fact happened                         |
| hasItem   | Checks if the player has item with given name             |
| hasMoney  | Checks if the player has *at least* given amount of money |

Conditions can be combined using `&&` and negated using `!`.

Examples:

```
if: fact arrivedHome
```

```
unless: hasItem Favorite Stopwatch
```

```
if: !hasMoney 100 && fact metWithJane
```

#### Macros

| Macro                | Description                                       |
|----------------------|---------------------------------------------------|
| addItem              | Adds 1 item with the given name                   |
| takeItem             | Takes 1 item with the given name                  |
| setFact              | Sets the given fact as true                       |
| addMoney             | Adds the given amount of money                    |
| takeMoney            | Takes the given amount of money                   |
| playMusic            | Plays the given music track (empty value to stop) |
| changeCharacterSpeed | Changes the character speed by given value        |

Examples:

```
addItem: Favorite Stopwatch
setFact: arrivedHome
addMoney: 100
```

#### Paragraphs

TODO

```
[h1]
[center]
[continue]
```

```
[effect typing 2s]
[continue]

[after 2s]
[continue]
```

#### Links

Links can appear anywhere in the text, but they will always show at the bottom of the dialog,
in the same order.

They must appear in a separate line, with double square brackets around the link text (`[[ ]]`).
The `> ` prefix is optional.

The link text is what's displayed and also the target passage (the next passage to be displayed).
You can use a different label using the `->` or `<-` syntax.

Examples:

```
> [[Ask him about Pete.]]
> [["What?"->Continue the conversation.]]
> [[Kitchen<-Take the door on the right.]]
```

Links use conditions in the same way as paragraphs.

Examples:

```
[if hasMoney 100]
> [[Buy the stopwatch ($1)]]
[continue]
```

Links can have tags.

| Tag   | Description                                                                     |
|-------|---------------------------------------------------------------------------------|
| back  | Marks the link as end of "subdialog", so it's later marked as visited           |
| exit  | Closes the dialog                                                               |
| level | Opens the level. Has a value after colon (`:`): level name and entrypoint index |

Examples:

```
> {back} [["I've heard enough."->Shopkeeper]]
```

```
> {exit} [["Bye for now."]]
```

```
> {level:bedroom,0} [[Open the door.]]
```

#### Tags

You can add tags in the Twine editor on the passage level.

| Tag  | Description                           |
|------|---------------------------------------|
| once | The passage can be visited only once. |

### Levels

| Object type |
|-------------|
| collider    |
| entrypoint  |
| fadepoint   |
| limits      |
| object      |
| poi         |
| trigger     |
