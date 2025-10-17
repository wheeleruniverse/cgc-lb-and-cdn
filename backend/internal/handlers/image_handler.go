package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"cgc-lb-and-cdn-backend/internal/agents"
	"cgc-lb-and-cdn-backend/internal/models"
	"cgc-lb-and-cdn-backend/internal/storage"
	"cgc-lb-and-cdn-backend/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// prompts contains 500 unique image generation prompts
var prompts = []string{
	// Animals with Jobs
	"A koala wearing a tiny firefighter's helmet, climbing a ladder to rescue a cat from a tree.",
	"An elegant giraffe working as a professional violinist in a concert hall.",
	"A team of squirrels in construction vests, building a miniature skyscraper out of acorns.",
	"A hamster dressed as a mad scientist, running on a wheel that powers a small laser.",
	"A chameleon wearing a detective trench coat, blending into a cluttered bookshelf.",
	"A group of penguins in suits, presenting a quarterly report in a chilly boardroom.",
	"An octopus barista, expertly making lattes with eight arms at a bustling coffee shop.",
	"A wise owl in a professor's cap and gown, teaching a class of baby birds.",
	"A majestic lion working as a librarian, quietly shelving books with a stern but fair expression.",
	"A golden retriever wearing a hard hat and safety goggles, inspecting a construction site.",

	// Fantasy and Mythical
	"A friendly dragon, meticulously tending a garden of glowing, fantastical flowers.",
	"A whimsical gnome architect, designing a house carved from a giant mushroom.",
	"A griffin delivering mail to a tiny floating village in the sky.",
	"An elegant fairy librarian, organizing a library of books with pages made of autumn leaves.",
	"A family of yetis having a picnic on a snowy mountain peak.",
	"A benevolent kraken playing chess against a tiny sailboat on a calm sea.",
	"A unicorn in an enchanted forest, serving tea to woodland creatures.",
	"A phoenix made of flowing molten glass, taking flight from a volcanic crater.",
	"A mischievous satyr playing a pan flute that makes flowers instantly bloom.",
	"A wise wizard using a sparkling wand to bake a cake for a child's birthday.",

	// Sci-Fi and Futuristic
	"A retro-futuristic robot, serving a cup of coffee at a space diner.",
	"A bustling city where all the buildings are giant, glowing crystals.",
	"A friendly alien tourist taking a selfie in front of the Eiffel Tower.",
	"An astronaut in a classic spacesuit, fishing on a distant, peaceful planet.",
	"A hovercraft shaped like a giant loaf of bread, delivering sandwiches.",
	"A futuristic food truck selling \"stardust tacos\" in a neon-lit alleyway.",
	"A cyborg with a heart of gold, building a birdhouse in a lush garden.",
	"A family of robots on a road trip through a galaxy of colorful gas clouds.",
	"A high-tech space port where ships are docked like planes at an airport.",
	"A giant robot, holding a sign that says \"Please Recycle.\"",

	// Nature and Outdoors
	"A friendly-looking squirrel riding a unicycle on a path through an autumn forest.",
	"A family of turtles enjoying a leisurely boat ride on a lily-pad pond.",
	"A whimsical treehouse with a spiral staircase and glowing lanterns.",
	"A vibrant field of sunflowers that turn to face the sun in a synchronized dance.",
	"A calm river flowing through a canyon made of oversized, colorful geodes.",
	"A curious fox peeking out from behind a vibrant, glowing waterfall.",
	"A bustling beehive that looks like a miniature, bustling city.",
	"A peaceful cottage nestled among giant, cloud-like lavender bushes.",
	"A garden where all the plants are made of different types of candy.",
	"A majestic whale with a glowing constellation pattern on its back, swimming in a starry ocean.",

	// Objects with Personality
	"A grumpy old toaster, trying to make the perfect toast.",
	"A friendly, smiling cloud wearing a top hat and a monocle.",
	"A vintage camera with a single, expressive eye, capturing a happy moment.",
	"A pencil and eraser, walking hand-in-hand down a winding road of a sketchbook.",
	"A happy, bouncing red ball, leaving a trail of rainbows.",
	"A wise old teacup, sitting on a shelf, with a small steam cloud that tells stories.",
	"A pair of mismatched socks, finally reunited after a long journey.",
	"A stack of books, happily celebrating the first day of school.",
	"A set of garden tools having a friendly conversation in a shed.",
	"A tiny, glowing lightbulb having a brilliant idea.",

	// Food and Drink
	"A sushi chef, meticulously preparing a plate of sushi on a tiny, detailed stage.",
	"A smiling ice cream cone, melting happily in the summer sun.",
	"A family of pastries, having a tea party in a whimsical kitchen.",
	"A friendly bowl of ramen, with noodles that look like tiny, smiling worms.",
	"A happy, bubbly soda can, playing a video game.",
	"A slice of pizza, wearing a tiny superhero cape, ready to save the day.",
	"A group of vegetables, forming a band and playing instruments made of kitchen utensils.",
	"A cheerful cup of hot chocolate, with marshmallows that look like fluffy clouds.",
	"A tiny, adventurous strawberry, scaling a mountain of whipped cream.",
	"A taco, dressed as a detective, investigating a case of missing salsa.",

	// Transportation and Vehicles
	"A hot air balloon shaped like a giant ice cream sundae, floating over a city.",
	"A whimsical train with a teapot for a boiler, traveling through a teacup landscape.",
	"A tiny submarine, exploring a beautiful coral reef made of gemstones.",
	"A friendly, old-fashioned bicycle, with a flower basket full of sunshine.",
	"A spaceship shaped like a rubber duck, flying through a starry, cosmic bath.",
	"A vintage car with a garden growing in its trunk.",
	"A cheerful sailboat with a sail made of patchwork quilts.",
	"A hot dog vendor cart, being pulled by a team of tiny, happy sausages.",
	"A cheerful, red fire truck with a hose that sprays confetti.",
	"A sleek, futuristic racing car, driving on a track made of light.",

	// Abstract and Surreal
	"A landscape where the sky is a swirling vortex of vibrant, pastel colors.",
	"A whimsical clock with hands that point to feelings instead of hours.",
	"A staircase that leads to a door opening into a sky full of fish.",
	"A single, glowing feather, floating in a room filled with giant, sparkling bubbles.",
	"A tree with roots that are also the branches, creating a perfect circle.",
	"A serene lake that reflects a different, fantastical world.",
	"A quiet room where all the furniture is made of different clouds.",
	"A majestic mountain range made of neatly folded blankets.",
	"A bookshelf where the books are filled with liquid light.",
	"A city skyline where buildings are made of giant, interlocking gears.",

	// Sports and Hobbies
	"A group of teacups, playing a game of miniature golf.",
	"A family of teddy bears, having a grand picnic and playing frisbee.",
	"A happy, colorful robot, painting a masterpiece on an oversized canvas.",
	"A trio of cats, expertly playing an intense game of chess.",
	"A cheerful, bouncing basketball, practicing its free throws.",
	"A group of friendly monsters, having a dance-off in a disco.",
	"A tiny, adventurous snail, hiking up a giant mountain.",
	"A family of garden gnomes, having a friendly race on their tricycles.",
	"A smiling, happy sun, playing hide-and-seek with the moon.",
	"A friendly ghost, learning to play the guitar.",

	// Everyday Life with a Twist
	"A busy city street where the cars are tiny, flying hot dogs.",
	"A serene park bench where a pigeon and a squirrel are reading a newspaper together.",
	"A cozy living room where a dog and a cat are sharing popcorn and watching a movie.",
	"A bustling laundromat where the washing machines are giant, smiling fishbowls.",
	"A family of socks, hanging out on a clothesline and telling jokes.",
	"A happy, bubbling bathtub, full of bubbles shaped like stars.",
	"A quiet library where the books float down to you on a magical breeze.",
	"A busy office where all the computers are powered by tiny, industrious gnomes.",
	"A peaceful night sky where the stars are actually tiny, glowing origami stars.",
	"A sunny day at the beach, where the sandcastles are made of colorful jelly.",

	// Additional Prompts (added 2025-10-17)
	"A distinguished polar bear working as a sommelier in an upscale restaurant.",
	"A clever raccoon operating a sophisticated recycling sorting facility.",
	"A patient sloth working as a meditation instructor at a wellness center.",
	"A nimble ferret conducting an orchestra of woodland creatures.",
	"A brave hedgehog serving as a night watchman with a tiny flashlight.",
	"A sophisticated peacock modeling haute couture on a glamorous runway.",
	"A hardworking beaver architect designing an eco-friendly dam community.",
	"A talented otter teaching a pottery class by the riverside.",
	"A wise tortoise working as a museum curator of ancient artifacts.",
	"A energetic chipmunk running a bustling farmers market stand.",
	"A graceful swan ballet instructor teaching baby ducklings to dance.",
	"A clever crow operating a lost-and-found service in the park.",
	"A friendly capybara working as a spa attendant at a hot spring.",
	"A determined mole engineer building an underground metro system.",
	"A cheerful puffin delivering newspapers to coastal villages.",
	"A skilled kangaroo working as a personal trainer at a gym.",
	"A gentle manatee lifeguard watching over swimmers at a tropical beach.",
	"A playful red panda working as a tea sommelier in a mountain café.",
	"A diligent ant foreman managing a construction site with blueprints.",
	"A proud peacock working as an art gallery docent.",
	"A curious lemur scientist conducting experiments in a jungle laboratory.",
	"A focused eagle air traffic controller at a busy airport.",
	"A motherly hen running a daycare center for baby birds.",
	"A sophisticated alpaca working as a luxury textile designer.",
	"A talented mockingbird impersonating famous singers on stage.",
	"A wise elephant historian writing memoirs in a study.",
	"A nimble gecko window washer scaling a tall skyscraper.",
	"A cheerful dolphin tour guide leading underwater sightseeing trips.",
	"A determined honey badger working as a treasure hunter.",
	"A patient spider web designer creating intricate digital networks.",
	"A jolly walrus ice sculptor creating masterpieces in the Arctic.",
	"A meticulous beetle jeweler crafting tiny, shimmering accessories.",
	"A regal peacock working as a luxury hotel concierge.",
	"A clever octopus locksmith with eight tools at once.",
	"A brave firefly lighthouse keeper guiding ships at night.",
	"A gentle giant panda working as a bamboo forest ranger.",
	"A energetic hummingbird barista making specialty nectar drinks.",
	"A wise owl judge presiding over a forest court.",
	"A skilled archer fish working as a professional basketball player.",
	"A talented chameleon makeup artist backstage at a theater.",
	"A dedicated bloodhound private investigator following a case.",
	"A cheerful sea otter sushi chef preparing fresh seafood.",
	"A sophisticated flamingo fashion designer sketching pink designs.",
	"A hardworking meerkat security guard monitoring surveillance cameras.",
	"A graceful seahorse ballet dancer performing underwater.",
	"A patient tortoise taxi driver navigating city streets slowly.",
	"A brave mongoose firefighter rescuing animals from danger.",
	"A talented parrot translator working at the United Nations.",
	"A diligent hamster accountant running on a calculator wheel.",
	"A friendly narwhal dentist with a natural unicorn horn tool.",
	"A gentle centaur blacksmith forging magical horseshoes in a misty forge.",
	"A mischievous leprechaun banker counting gold coins in a rainbow vault.",
	"A elegant mermaid concert pianist playing in an underwater amphitheater.",
	"A noble pegasus mail carrier delivering cloud letters across the sky.",
	"A wise sphinx librarian guarding riddles written in ancient scrolls.",
	"A playful pixie gardener tending to miniature enchanted toadstools.",
	"A brave minotaur maze designer creating elaborate labyrinth puzzles.",
	"A serene nymph watercolorist painting by a crystalline stream.",
	"A mysterious banshee opera singer performing in a haunted theater.",
	"A friendly troll bridge inspector maintaining crossing safety.",
	"A magical kitsune illusionist performing at a mystical circus.",
	"A gentle giant working as a cloud shepherd in the sky.",
	"A clever goblin inventor tinkering with steampunk contraptions.",
	"A ethereal will-o'-wisp tour guide leading travelers through foggy swamps.",
	"A majestic thunderbird weather forecaster predicting storms.",
	"A mischievous brownie chef baking midnight treats in a cottage kitchen.",
	"A ancient dryad botanist studying magical tree species.",
	"A graceful sylph aerial acrobat dancing on wind currents.",
	"A mysterious grim working as a guardian of crossroads.",
	"A cheerful gnome watchmaker crafting tiny mechanical timepieces.",
	"A noble gryphon knight guarding a castle's treasure tower.",
	"A wise oracle fortune teller reading crystal balls in a tent.",
	"A playful imp practical joker setting up harmless magical pranks.",
	"A serene undine water purification specialist at a sacred spring.",
	"A brave valkyrie warrior training new heroes in Valhalla.",
	"A mysterious changeling actor transforming for different roles.",
	"A friendly hobbit chef running a cozy countryside inn.",
	"A ancient basilisk sculptor creating stone statues with a glance.",
	"A graceful selkie marine biologist studying coastal ecosystems.",
	"A clever djinn wish consultant helping clients word requests carefully.",
	"A ethereal ghost historian documenting haunted house histories.",
	"A playful faun musician playing pan pipes in moonlit glades.",
	"A wise elder ent arborist caring for ancient forest groves.",
	"A mysterious vampire sommelier curating rare vintage wines.",
	"A cheerful cupid matchmaker arranging perfect love connections.",
	"A noble gargoyle architect perched atop Gothic cathedrals.",
	"A gentle yeti meteorologist forecasting mountain weather patterns.",
	"A clever roc pilot transporting cargo across impossible distances.",
	"A ancient chimera veterinarian with expertise in hybrid creatures.",
	"A graceful harpy messenger delivering urgent scrolls by air.",
	"A mysterious werewolf nightshift security guard under moonlight.",
	"A friendly bogeyman closet organizer helping kids face their fears.",
	"A wise crone herbalist brewing healing potions in a forest cabin.",
	"A playful satyr vintner stomping grapes in a hillside vineyard.",
	"A ethereal banshee grief counselor helping souls find peace.",
	"A brave Amazon warrior teaching self-defense classes.",
	"A mysterious medusa hairstylist creating stunning stone sculptures.",
	"A cheerful tooth fairy dental hygienist on night rounds.",
	"A ancient phoenix life coach helping others rise from ashes.",
	"A noble Pegasus flight instructor teaching young winged horses.",
	"A android chef preparing molecular gastronomy in a space station.",
	"A time traveler historian documenting alternate timelines in a chrono-lab.",
	"A holographic pop star performing concerts across multiple dimensions.",
	"A robot gardener cultivating hydroponic vegetables on Mars.",
	"A alien diplomat negotiating peace treaties between star systems.",
	"A cyborg athlete competing in zero-gravity Olympic games.",
	"A AI therapist providing emotional support to lonely astronauts.",
	"A quantum physicist cat studying Schrödinger's experiment from inside.",
	"A nanobots swarm working together to repair a damaged spaceship.",
	"A teleporter technician maintaining wormhole transit stations.",
	"A space miner extracting precious crystals from asteroid belts.",
	"A virtual reality designer creating immersive dream worlds.",
	"A genetic engineer cultivating bioluminescent forests on exoplanets.",
	"A plasma welder building the framework of a new space colony.",
	"A antimatter fuel specialist maintaining starship power cores.",
	"A exobiologist discovering new life forms in alien oceans.",
	"A drone swarm coordinator managing delivery logistics in a megacity.",
	"A cryogenic technician monitoring frozen colonists on a generation ship.",
	"A force field engineer protecting settlements from solar radiation.",
	"A dark matter researcher studying the invisible universe.",
	"A terraforming specialist converting barren worlds into habitable paradises.",
	"A neural interface designer linking minds to advanced computers.",
	"A solar sail navigator charting courses through interstellar space.",
	"A gravity generator mechanic keeping space stations properly oriented.",
	"A clone coordinator managing duplicate work shifts on moon bases.",
	"A photon artist painting with pure light beams.",
	"A dimensional rift sealer preventing multiverse paradoxes.",
	"A bioship pilot merging consciousness with a living spacecraft.",
	"A electromagnetic pulse shieldsmith protecting cities from tech attacks.",
	"A memory backup specialist digitizing consciousness for immortality.",
	"A asteroid farmer growing crops in spinning rock gardens.",
	"A plasma storm chaser studying stellar weather phenomena.",
	"A transdimensional postal worker delivering packages across realities.",
	"A laser sculptor carving intricate designs in floating metal.",
	"A space debris collector cleaning up orbital junk with magnetic nets.",
	"A singularity researcher studying black hole event horizons safely.",
	"A stardust harvester collecting cosmic particles for manufacturing.",
	"A tachyon communicator enabling faster-than-light messaging.",
	"A vacuum energy tapper drawing power from empty space.",
	"A cosmic string cartographer mapping the universe's fundamental structure.",
	"A warp bubble technician maintaining faster-than-light engines.",
	"A exosuit designer creating adaptive armor for alien environments.",
	"A stellar nursery observer watching new stars being born.",
	"A hyperspace navigator plotting routes through folded space.",
	"A antimatter containment specialist preventing catastrophic explosions.",
	"A chronolock engineer ensuring time flows properly in relativistic travel.",
	"A megastructure architect designing Dyson spheres around suns.",
	"A quantum entanglement communicator maintaining instant galactic networks.",
	"A synthetic consciousness ethicist evaluating AI sentience rights.",
	"A universal translator linguist decoding alien languages instantly.",
	"A wise old redwood tree with a face in its bark telling ancient stories.",
	"A family of mushrooms glowing softly in a enchanted midnight forest.",
	"A crystal cave with stalactites that chime like wind bells.",
	"A mountain peak where clouds gather to share weather gossip.",
	"A coral reef city bustling with colorful fish traffic.",
	"A desert oasis where cacti bloom with rainbow flowers.",
	"A bamboo forest where pandas practice martial arts.",
	"A tidal pool reflecting an entire miniature ocean ecosystem.",
	"A volcanic island with friendly lava flows that wave hello.",
	"A glacier carving intricate ice sculptures as it slowly moves.",
	"A kelp forest swaying in underwater currents like a green ballet.",
	"A canyon painted in layers of geological time.",
	"A geyser that erupts on a precise schedule like a natural clock.",
	"A meadow where butterflies migrate in kaleidoscope formations.",
	"A mangrove maze where roots create natural tunnels.",
	"A salt flat reflecting the sky like Earth's largest mirror.",
	"A Northern Lights dancing above a peaceful Arctic landscape.",
	"A hot spring terraces cascading down a mountainside in pastel colors.",
	"A ancient grove where trees have grown into natural archways.",
	"A sand dune field singing in harmonic tones as wind passes.",
	"A bioluminescent bay glowing blue with plankton at night.",
	"A petrified forest where ancient trees turned to stone.",
	"A moss garden covering rocks in every shade of green.",
	"A tide coming in to reveal a hidden beach cave.",
	"A waterfall cascading through a rainbow in perpetual mist.",
	"A alpine meadow blooming with wildflowers in concentric circles.",
	"A ancient baobab tree with a hollow trunk large enough for a room.",
	"A limestone formations creating a natural stone bridge.",
	"A river delta branching into fractal patterns from above.",
	"A redwood canopy where entire ecosystems exist hundreds of feet up.",
	"A sandstone arch framing a desert sunset perfectly.",
	"A field of cattails swaying in synchronization with the breeze.",
	"A frost covering autumn leaves in delicate crystalline patterns.",
	"A underground river flowing through glowing crystal caverns.",
	"A prairie where grass waves like a golden ocean.",
	"A wetland where birds and frogs create a evening symphony.",
	"A rocky coastline where tide pools form natural aquariums.",
	"A valley where morning fog settles like a fluffy blanket.",
	"A sakura tree shedding pink petals in a gentle spring breeze.",
	"A lagoon where fresh and saltwater create unique ecosystems.",
	"A thunderstorm rolling across plains with dramatic lightning.",
	"A autumn forest floor carpeted in colorful fallen leaves.",
	"A snow-covered pine forest silent and peaceful.",
	"A natural hot spring in the middle of a snowy landscape.",
	"A flower field where bees dance from bloom to bloom.",
	"A jungle canopy where exotic birds create a living rainbow.",
	"A mountain lake so clear you can see to the bottom.",
	"A redwood nurse log sprouting new trees from its decomposing form.",
	"A coastal cliff where seabirds nest in natural alcoves.",
	"A monsoon creating temporary waterfalls on every cliff face.",
	"A loyal alarm clock that apologizes for waking you up.",
	"A ambitious staircase dreaming of becoming an escalator.",
	"A nervous printer afraid of running out of ink.",
	"A proud refrigerator showing off its organized interior.",
	"A friendly doorbell that sings instead of rings.",
	"A wise old dictionary sharing the origins of words.",
	"A playful yo-yo showing off new tricks.",
	"A tired coffee maker working the morning shift.",
	"A cheerful spatula flipping pancakes with enthusiasm.",
	"A sophisticated wine glass discussing proper aeration.",
	"A determined mop cleaning up after a party.",
	"A artistic paint brush creating a self-portrait.",
	"A musical keyboard playing a happy tune by itself.",
	"A cozy blanket wrapping itself around someone cold.",
	"A adventurous kite soaring higher than ever before.",
	"A precise metronome keeping perfect time.",
	"A helpful flashlight guiding someone through darkness.",
	"A vintage typewriter writing poetry late at night.",
	"A happy watering can nurturing a window garden.",
	"A philosophical hourglass contemplating the passage of time.",
	"A brave umbrella standing up to a fierce storm.",
	"A friendly mailbox excited to receive letters.",
	"A curious telescope gazing at distant galaxies.",
	"A hardworking broom sweeping up stardust.",
	"A talented saxophone playing smooth jazz.",
	"A warm fireplace crackling contentedly.",
	"A loyal backpack carrying treasures from adventures.",
	"A wise compass always pointing toward true north.",
	"A cheerful sunflower seeds ready to grow.",
	"A sophisticated fountain pen writing elegant calligraphy.",
	"A brave candle illuminating a dark room.",
	"A patient hourglass marking meditation sessions.",
	"A musical triangle waiting for its moment to shine.",
	"A helpful bookmark saving someone's place in an epic story.",
	"A proud trophy recounting the victory it represents.",
	"A cozy hammock swaying gently in the breeze.",
	"A determined zipper trying to close a overstuffed suitcase.",
	"A friendly welcome mat greeting visitors warmly.",
	"A artistic crayon box showcasing a rainbow of colors.",
	"A loyal dog collar remembering wonderful walks.",
	"A precise ruler measuring life's little details.",
	"A cheerful wind chime creating a peaceful melody.",
	"A wise old grandfather clock keeping family time.",
	"A adventurous paper airplane soaring across a classroom.",
	"A helpful sticky note reminding someone of something important.",
	"A talented kazoo humming a cheerful tune.",
	"A warm tea kettle whistling a happy song.",
	"A friendly pillow supporting sweet dreams.",
	"A brave night light keeping shadows at bay.",
	"A sophisticated monocle examining the finer details.",
	"A sophisticated espresso shot giving a morning pep talk.",
	"A cheerful donut with sprinkles celebrating being someone's favorite.",
	"A wise aged cheese discussing its complex flavor profile.",
	"A brave jalapeño pepper bragging about its heat level.",
	"A friendly baguette fresh from a Parisian bakery.",
	"A elegant champagne bottle popping for a celebration.",
	"A adventurous curry dish from a bustling street market.",
	"A cozy pot of soup simmering with love and herbs.",
	"A proud wedding cake standing tall with multiple tiers.",
	"A playful popcorn kernels popping in excitement.",
	"A sophisticated truffle sharing its earthy secrets.",
	"A cheerful breakfast burrito wrapped up and ready to go.",
	"A artistic sushi roll arranged like a work of art.",
	"A warm croissant fresh and flaky with butter.",
	"A refreshing lemonade with perfect sweet-tart balance.",
	"A noble roasted turkey at the center of a feast.",
	"A happy dumpling family steaming in a bamboo basket.",
	"A adventurous pho bowl with aromatic herbs and spices.",
	"A elegant macaron tower in pastel rainbow colors.",
	"A brave wasabi warning diners of its intense power.",
	"A friendly apple pie cooling on a windowsill.",
	"A sophisticated aged wine discussing its vintage year.",
	"A cheerful bubble tea with tapioca pearls bouncing.",
	"A wise sourdough starter centuries old and still active.",
	"A playful cotton candy cloud on a stick.",
	"A determined espresso machine working through morning rush.",
	"A elegant tiramisu layered to perfection.",
	"A adventurous kimchi fermenting with probiotic pride.",
	"A cozy hot toddy warming someone on a cold night.",
	"A proud paella pan filled with saffron rice and seafood.",
	"A cheerful waffle with butter and syrup rivers.",
	"A sophisticated caviar discussing luxury dining.",
	"A friendly miso soup starting the day right.",
	"A artistic gelato display in an Italian shop window.",
	"A brave ghost pepper challenging spice enthusiasts.",
	"A elegant crème brûlée with a perfectly torched top.",
	"A adventurous shawarma spinning on a vertical rotisserie.",
	"A wise aged balsamic vinegar from Modena.",
	"A cheerful churro dusted with cinnamon sugar.",
	"A sophisticated single-origin coffee explaining its terroir.",
	"A friendly pretzel twisted into a perfect knot.",
	"A playful jelly beans in a rainbow assortment.",
	"A determined pressure cooker making a quick meal.",
	"A elegant Napoleon pastry with crispy layers.",
	"A cheerful breakfast cereal providing morning nutrition.",
	"A sophisticated foie gras debating culinary ethics.",
	"A adventurous durian fruit with a controversial reputation.",
	"A warm apple cider spiced for autumn.",
	"A proud standing rib roast at a holiday dinner.",
	"A cheerful smoothie bowl topped with fresh fruit art.",
	"A vintage steam locomotive chugging through mountain passes.",
	"A friendly gondola gliding through Venetian canals.",
	"A brave rickshaw navigating busy Delhi streets.",
	"A elegant yacht sailing into a Mediterranean sunset.",
	"A cheerful double-decker bus touring London landmarks.",
	"A adventurous dog sled team racing across frozen tundra.",
	"A sophisticated bullet train gliding silently at high speed.",
	"A playful bumper car at a carnival fairground.",
	"A wise old lighthouse keeping ships safe for centuries.",
	"A determined snowplow clearing roads before dawn.",
	"A elegant horse-drawn carriage in Central Park.",
	"A brave icebreaker ship cutting through Arctic waters.",
	"A cheerful tuk-tuk weaving through Bangkok traffic.",
	"A adventurous cable car ascending a steep mountain.",
	"A sophisticated seaplane landing on a remote lake.",
	"A friendly milk truck making early morning deliveries.",
	"A playful zip line soaring over jungle canopy.",
	"A determined ambulance rushing through city streets.",
	"A elegant trolley car climbing San Francisco hills.",
	"A brave coast guard boat responding to emergencies.",
	"A cheerful ice cream truck playing nostalgic melodies.",
	"A adventurous paraglider riding thermal updrafts.",
	"A sophisticated limousine arriving at a red carpet event.",
	"A wise old drawbridge raising to let tall ships pass.",
	"A determined garbage truck completing its essential route.",
	"A elegant rickshaw decorated with colorful paintings.",
	"A brave helicopter landing on a mountain rescue mission.",
	"A cheerful paddleboat shaped like a giant swan.",
	"A adventurous hang glider catching perfect wind.",
	"A sophisticated hovercraft crossing from land to water.",
	"A friendly postal truck delivering mail to rural areas.",
	"A playful roller coaster climbing to its highest peak.",
	"A determined snow groomer preparing perfect ski slopes.",
	"A elegant junk boat with distinctive red sails.",
	"A brave lifeboat launching in rough seas.",
	"A cheerful carousel with hand-painted horses.",
	"A adventurous mountain bike tackling rough trails.",
	"A sophisticated private jet crossing continents.",
	"A wise old ferry connecting island communities.",
	"A determined tow truck rescuing stranded vehicles.",
	"A elegant sampan boat floating through floating markets.",
	"A brave snowmobile racing across frozen lakes.",
	"A cheerful segway tour rolling through historic districts.",
	"A adventurous hot rod at a vintage car show.",
	"A sophisticated catamaran sailing in tropical waters.",
	"A friendly school bus safely transporting children.",
	"A playful kiddie train circling a shopping mall.",
	"A determined cement mixer building new construction.",
	"A elegant canal boat navigating historic waterways.",
	"A brave ski lift carrying skiers up snowy peaks.",
	"A staircase that spirals into a sunset instead of a ceiling.",
	"A door that opens to different seasons each time you turn the knob.",
	"A mirror reflecting tomorrow instead of today.",
	"A piano where each key plays a different emotion.",
	"A telescope that shows the past instead of distant stars.",
	"A umbrella that rains upward into the sky.",
	"A bridge connecting two different paintings.",
	"A window showing a view from another planet.",
	"A hourglass where sand flows in both directions simultaneously.",
	"A compass pointing toward your heart's desire.",
	"A book where the words rearrange themselves to tell your story.",
	"A garden where memories grow as flowers.",
	"A candle whose flame is made of frozen ice.",
	"A shadow that exists without an object to cast it.",
	"A rainbow that curves into a perfect mathematical spiral.",
	"A cloud shaped like a question mark raining answers.",
	"A tree that grows light bulbs instead of fruit.",
	"A river that flows vertically up a mountain.",
	"A moon that changes phases based on your mood.",
	"A painting that changes scenes when you're not looking.",
	"A sculpture that casts a shadow of a completely different object.",
	"A carpet that shows footprints of people from the past.",
	"A fountain where water flows in geometric patterns.",
	"A chandelier made of suspended water droplets.",
	"A snow globe containing a miniature functioning city.",
	"A sundial that tells time in colors instead of numbers.",
	"A kaleidoscope showing infinite parallel universes.",
	"A wind that carries visible musical notes.",
	"A fog that reveals hidden truths as it lifts.",
	"A earthquake that only affects emotions, not buildings.",
	"A eclipse where the sun and moon trade places.",
	"A tide that brings in dreams instead of seashells.",
	"A thunderstorm that rains colors instead of water.",
	"A desert where sand dunes are actually frozen waves.",
	"A forest where trees are made of crystallized time.",
	"A sky where clouds form words in ancient languages.",
	"A ocean where waves create symphonies as they crash.",
	"A mountain whose peak touches the bottom of the sea.",
	"A valley where echoes arrive before the original sound.",
	"A cave where stalactites grow downward into stars.",
	"A meadow where grass blades are actually tiny antennae.",
	"A horizon that curves upward into the sky.",
	"A whirlpool that spins clockwise and counterclockwise simultaneously.",
	"A lighthouse beam that illuminates memories instead of sea.",
	"A butterfly with wings showing different realities.",
	"A spiderweb woven from moonbeams.",
	"A raindrop that falls upward into clouds.",
	"A stone that ripples like water when touched.",
	"A flame that casts darkness instead of light.",
	"A infinity symbol walking like a figure-eight creature.",
}

// ImageHandler handles image generation requests
type ImageHandler struct {
	orchestrator agents.OrchestratorAgent
	valkeyClient *storage.ValkeyClient
}

// NewImageHandler creates a new image handler
func NewImageHandler(orchestrator agents.OrchestratorAgent, valkeyClient *storage.ValkeyClient) *ImageHandler {
	return &ImageHandler{
		orchestrator: orchestrator,
		valkeyClient: valkeyClient,
	}
}

// getRandomPrompt returns a random prompt from the prompts list
func getRandomPrompt() string {
	return prompts[rand.Intn(len(prompts))]
}

// GenerateImage handles POST /generate requests
func (h *ImageHandler) GenerateImage(c *gin.Context) {
	var req models.ImageRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", map[string]string{
			"validation_error": err.Error(),
		})
		return
	}

	// Set request metadata
	requestID := uuid.New().String()
	pairID := uuid.New().String()
	req.RequestID = requestID
	req.PairID = pairID
	req.Timestamp = time.Now()

	// Use random prompt
	req.Prompt = getRandomPrompt()
	fmt.Printf("[INFO] Using random prompt: %s, pair_id: %s\n", req.Prompt, pairID)

	// Generate image pair (2 images: left and right)
	result, err := h.orchestrator.Execute(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Image generation failed", "GENERATION_FAILED", map[string]string{
			"error": err.Error(),
		})
		return
	}

	response, ok := result.(*models.ImageResponse)
	if !ok || len(response.Images) < 2 {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid response - need 2 images", "INVALID_RESPONSE", nil)
		return
	}

	// First image is "left" (index 0), second is "right" (index 1)
	leftImage := response.Images[0]
	rightImage := response.Images[1]
	timestamp := time.Now()

	// Store the pair in Valkey (simplified structure with pair-id only)
	if h.valkeyClient != nil {
		pair := &storage.ImagePair{
			PairID:    pairID,
			Prompt:    req.Prompt,
			Provider:  response.Provider,
			LeftURL:   leftImage.URL,
			RightURL:  rightImage.URL,
			Timestamp: timestamp,
		}

		if err := h.valkeyClient.StoreImagePair(c.Request.Context(), pair); err != nil {
			fmt.Printf("[ERROR] Failed to store image pair: %v\n", err)
			// Continue anyway - don't fail the request
		} else {
			fmt.Printf("[PAIR] Stored in Valkey - Pair: %s, Prompt: %s, Provider: %s\n", pairID, req.Prompt, response.Provider)
		}
	}

	// Return success response with both images
	utils.RespondWithSuccess(c, gin.H{
		"pair_id":     pairID,
		"prompt":      req.Prompt,
		"provider":    response.Provider,
		"left_image":  leftImage,
		"right_image": rightImage,
		"timestamp":   timestamp.Format(time.RFC3339),
	}, "Image pair generated successfully", map[string]string{
		"pair_id":    pairID,
		"request_id": requestID,
		"provider":   response.Provider,
	})
}

// GetProviderStatus handles GET /status requests
func (h *ImageHandler) GetProviderStatus(c *gin.Context) {
	// Check if quota refresh is requested
	refreshQuota := c.Query("refresh_quota") == "true"

	if refreshQuota {
		h.refreshAllProviderQuotas(c.Request.Context())
	}

	status := h.orchestrator.GetProviderStatus()

	utils.RespondWithSuccess(c, status, "Provider status retrieved", map[string]string{
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"quota_refreshed": fmt.Sprintf("%t", refreshQuota),
	})
}

// refreshAllProviderQuotas refreshes quota information for all providers
func (h *ImageHandler) refreshAllProviderQuotas(ctx context.Context) {
	providerStatus := h.orchestrator.GetProviderStatus()

	for providerName := range providerStatus {
		fmt.Printf("[STATUS] Refreshing quota for provider: %s\n", providerName)

		provider, exists := h.orchestrator.GetProvider(providerName)
		if !exists {
			fmt.Printf("[STATUS] Provider %s not found\n", providerName)
			continue
		}

		if err := provider.RefreshQuota(ctx); err != nil {
			fmt.Printf("[STATUS] Failed to refresh quota for %s: %v\n", providerName, err)
		} else {
			fmt.Printf("[STATUS] Successfully refreshed quota for %s\n", providerName)
		}
	}
}

// HealthCheck handles GET /health requests
func (h *ImageHandler) HealthCheck(c *gin.Context) {
	status := h.orchestrator.GetProviderStatus()

	// Check if at least one provider is available
	availableCount := 0
	totalCount := len(status)

	for _, providerStatus := range status {
		if providerStatus.Available {
			availableCount++
		}
	}

	healthStatus := "healthy"
	statusCode := http.StatusOK

	if availableCount == 0 {
		healthStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	} else if availableCount < totalCount {
		healthStatus = "degraded"
	}

	c.JSON(statusCode, gin.H{
		"status":              healthStatus,
		"available_providers": availableCount,
		"total_providers":     totalCount,
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
		"providers":           status,
	})
}

// GetImagePair handles GET /images/pair requests
// Supports optional "exclude" query parameter with comma-separated pair IDs
// Supports optional "session_id" query parameter for session-based tracking
func (h *ImageHandler) GetImagePair(c *gin.Context) {
	if h.valkeyClient == nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Image pairs unavailable", "VALKEY_UNAVAILABLE", nil)
		return
	}

	// Get session ID from query parameter (optional)
	sessionID := c.Query("session_id")

	// Parse excluded pair IDs from query parameter
	excludedPairIDs := []string{}
	if excludeParam := c.Query("exclude"); excludeParam != "" {
		excludedPairIDs = strings.Split(excludeParam, ",")
		// Trim whitespace from each ID
		for i, id := range excludedPairIDs {
			excludedPairIDs[i] = strings.TrimSpace(id)
		}
	}

	// Get random pair from Valkey
	var pair *storage.ImagePair
	var err error

	if sessionID != "" {
		// Use session-based tracking to avoid showing same images to same user
		pair, err = h.valkeyClient.GetRandomImagePairForSession(c.Request.Context(), sessionID, excludedPairIDs)
	} else {
		// Fallback to original behavior (only use explicit exclusions)
		pair, err = h.valkeyClient.GetRandomImagePair(c.Request.Context(), excludedPairIDs)
	}
	if err != nil {
		// Check if it's an empty database (no pairs available yet)
		if strings.Contains(err.Error(), "no pairs available") {
			utils.RespondWithError(c, http.StatusNotFound, "No image pairs available yet. Images are being generated in the background - please check back in a few moments!", "NO_PAIRS_YET", map[string]string{
				"suggestion": "Try generating a new pair or wait for automatic generation",
			})
			return
		}

		// Check if all pairs have been voted on
		if strings.Contains(err.Error(), "no unvoted pairs available") {
			utils.RespondWithError(c, http.StatusNotFound, "You've voted on all available pairs! Great job! 🎉", "ALL_PAIRS_VOTED", map[string]string{
				"suggestion": "Check back later for new images",
			})
			return
		}

		// Other errors
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get random image pair", "PAIR_UNAVAILABLE", map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Simplified response: no duplicate data
	response := models.ImagePairResponse{
		PairID:   pair.PairID,
		Prompt:   pair.Prompt,
		Provider: pair.Provider,
		LeftURL:  pair.LeftURL,
		RightURL: pair.RightURL,
	}

	utils.RespondWithSuccess(c, response, "Image pair retrieved successfully", nil)
}

// SubmitRating handles POST /images/rate requests
func (h *ImageHandler) SubmitRating(c *gin.Context) {
	var req models.ComparisonRatingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", map[string]string{
			"validation_error": err.Error(),
		})
		return
	}

	// Validate winner value
	if req.Winner != "left" && req.Winner != "right" {
		utils.RespondWithError(c, http.StatusBadRequest, "Winner must be 'left' or 'right'", "INVALID_WINNER", nil)
		return
	}

	// Store vote in Valkey
	if h.valkeyClient != nil {
		// Fetch the image pair to get provider and prompt information
		pair, err := h.valkeyClient.GetImagePairByID(c.Request.Context(), req.PairID)
		if err != nil {
			// Pair not found - still record the vote but without provider/prompt
			fmt.Printf("[WARN] Could not fetch pair %s for vote metadata: %v\n", req.PairID, err)
		}

		vote := &storage.Vote{
			PairID:   req.PairID,
			Winner:   req.Winner,
			Provider: "",
			Prompt:   "",
		}

		// Add provider and prompt if we found the pair
		if pair != nil {
			vote.Provider = pair.Provider
			vote.Prompt = pair.Prompt
		}

		if err := h.valkeyClient.RecordVote(c.Request.Context(), vote); err != nil {
			fmt.Printf("[ERROR] Failed to record vote in Valkey: %v\n", err)
			// Continue anyway - don't fail the request if Valkey is down
		} else {
			fmt.Printf("[VOTE] Recorded in Valkey - Pair: %s, Winner: %s, Provider: %s\n", req.PairID, req.Winner, vote.Provider)
		}
	}

	response := models.ComparisonRatingResponse{
		Success:   true,
		PairID:    req.PairID,
		Winner:    req.Winner,
		Message:   "Rating submitted successfully",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	utils.RespondWithSuccess(c, response, "Rating submitted successfully", map[string]string{
		"pair_id": req.PairID,
		"winner":  req.Winner,
	})
}

// GetStatistics handles GET /statistics requests
func (h *ImageHandler) GetStatistics(c *gin.Context) {
	if h.valkeyClient == nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Statistics unavailable", "VALKEY_UNAVAILABLE", nil)
		return
	}

	totalVotes, err := h.valkeyClient.GetTotalVotes(c.Request.Context())
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get total votes", "STATISTICS_ERROR", map[string]string{
			"error": err.Error(),
		})
		return
	}

	sideWins, err := h.valkeyClient.GetSideWins(c.Request.Context())
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get side wins", "STATISTICS_ERROR", map[string]string{
			"error": err.Error(),
		})
		return
	}

	utils.RespondWithSuccess(c, gin.H{
		"total_votes": totalVotes,
		"side_wins":   sideWins,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}, "Statistics retrieved successfully", nil)
}

// GetWinners handles GET /images/winners requests
func (h *ImageHandler) GetWinners(c *gin.Context) {
	if h.valkeyClient == nil {
		utils.RespondWithError(c, http.StatusServiceUnavailable, "Winners unavailable", "VALKEY_UNAVAILABLE", nil)
		return
	}

	// Get side parameter (default to "left")
	side := c.DefaultQuery("side", "left")
	if side != "left" && side != "right" {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid side parameter", "INVALID_SIDE", map[string]string{
			"side":    side,
			"allowed": "left, right",
		})
		return
	}

	winningPairs, err := h.valkeyClient.GetWinningImages(c.Request.Context(), side)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get winners", "WINNERS_ERROR", map[string]string{
			"error": err.Error(),
			"side":  side,
		})
		return
	}

	// Transform to response format
	type WinnerImage struct {
		ImageURL  string `json:"image_url"`
		Prompt    string `json:"prompt"`
		Provider  string `json:"provider"`
		PairID    string `json:"pair_id"`
		Timestamp string `json:"timestamp"`
		VoteCount int64  `json:"vote_count"`
	}

	var winners []WinnerImage
	for _, pair := range winningPairs {
		imageURL := pair.LeftURL
		if side == "right" {
			imageURL = pair.RightURL
		}

		winners = append(winners, WinnerImage{
			ImageURL:  imageURL,
			Prompt:    pair.Prompt,
			Provider:  pair.Provider,
			PairID:    pair.PairID,
			Timestamp: pair.Timestamp.Format(time.RFC3339),
			VoteCount: pair.VoteCount,
		})
	}

	utils.RespondWithSuccess(c, gin.H{
		"winners":   winners,
		"count":     len(winners),
		"side":      side,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, fmt.Sprintf("%s winners retrieved successfully", strings.Title(side)), nil)
}
