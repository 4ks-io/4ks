package fetcher

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/google/go-cmp/cmp"
)

var serverIndexResponse = []byte("hello world\n")

func newTestHTTPServer(dat []byte) *httptest.Server {
	srv := newUnstartedTestServer(dat)

	// create a listener with the desired port.
	l, err := net.Listen("tcp", "127.0.0.1:44245")
	if err != nil {
		panic(err)
	}

	// NewUnstartedServer creates a listener. Close that listener and replace with new
	srv.Listener.Close()
	srv.Listener = l

	srv.Start()
	return srv
}

func newUnstartedTestServer(dat []byte) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/recipe", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(dat)
	})

	return httptest.NewUnstartedServer(mux)
}

const mockURL = "http://127.0.0.1:44245/recipe"

var testCases = map[string]struct {
	Callbacks int
	Skip      bool
	Recipe    Recipe
}{
	"bigoven": {
		Skip:      false,
		Callbacks: 1,
		Recipe: Recipe{
			SourceURL: mockURL,
			Title:     "Lemony Chicken and Rice Soup",
			Ingredients: []string{
				"1 tablespoon extra-virgin olive oil",
				"1 medium onion chopped",
				"4 celery stalks trimmed and sliced",
				"4 carrots sliced into rounds",
				"2 boneless, skinless chicken breasts",
				"8 cups chicken stock homemade or store-bought",
				"1 teaspoon kosher salt",
				"1/4 teaspoon black pepper",
				"1 cup long-grain white rice",
				"1/4 cup freshly squeezed lemon juice about 1 lemon",
				"1 teaspoon finely grated lemon zest optional",
				"1/4 cup fresh dill (or 1 tablespoon dried) roughly chopped",
				"1/4 cup fresh parsley (or 1 tablespoon dried) roughly chopped",
				"2 tablespoons unsalted butter",
			},
			Instructions: []string{
				"Heat olive oil in a heavy-bottomed Dutch oven set over medium-high heat.",
				"Add onion, celery, and carrots and sauté until the vegetables begin to soften, about 4 minutes.",
				"Add chicken stock, salt, and pepper and bring mixture to a boil. You can cover the pot to help it reach a boil faster.",
				"Sprinkle chicken breasts with salt and pepper. Once stock is boiling, add the chicken breasts and reduce heat to medium-low to simmer. Cover the pot with a lid and let the chicken poach for about 15 minutes.",
				"Use a meat thermometer to check the internal temperature of the chicken around the 15-minute mark. The chicken is cooked when it reaches 165 degrees Fahrenheit. The exact time to cook the chicken will depend on the size of the breasts.",
				"Once the chicken has reached an internal temperature of 165 degrees Fahrenheit, remove it from the pot.",
				"Add rice to the pot and cover again. Let simmer for 15 minutes more.",
				"While the rice cooks, shred the chicken breasts.",
				"When the rice has cooked through, return the shredded chicken to the pot. Stir in the lemon juice and zest, if using, dill, parsley, and butter. Serve once butter has melted.",
			},
		},
	},
	"minimalistbaker": {
		Skip:      false,
		Callbacks: 1,
		Recipe: Recipe{
			SourceURL: mockURL,
			Title:     "Creamy Vegan Tofu Cauliflower Curry (Korma-Inspired)",
			Ingredients: []string{
				"1 small head cauliflower, trimmed and cut into bite-sized florets",
				"1 (14 oz.) package  extra-firm tofu, torn into 1-inch pieces",
				"2 Tbsp olive or coconut oil",
				"1 Tbsp DIY curry powder ((or store-bought))",
				"1/2 tsp sea salt",
				"1 Tbsp olive or coconut oil",
				"1 medium onion, diced ((1 onion yields ~2 cups or 220 g))",
				"4 medium cloves garlic, minced",
				"1 Tbsp minced fresh ginger",
				"1 Tbsp DIY curry powder ((or store-bought))",
				"1 Tbsp tomato paste",
				"3/4 tsp sea salt",
				"2 ½ cups water ((DIVIDED))",
				"1/4 cup raw cashews",
				"1 Tbsp lemon juice ((optional))",
				"White or brown rice",
				"Coconut yogurt ((or other plain/unsweetened yogurt))",
				"Quick pickled onions",
				"Fresh cilantro",
			},
			Instructions: []string{
				"TOFU AND CAULIFLOWER: Preheat your oven to 425 degrees F (218 C) and line a baking sheet with parchment paper.",
				"Place the bite-sized cauliflower florets and 1-inch pieces of tofu onto the baking sheet. Drizzle with oil then sprinkle with curry powder and salt. Toss to thoroughly coat. Bake for 20-25 minutes, tossing halfway through. The cauliflower will be tender and golden brown on the outside and the tofu slightly crispy.",
				"KORMA SAUCE: Meanwhile, heat the oil in a large rimmed skillet or pot over medium heat. Once hot, add the onions, garlic, and ginger. Cook over medium heat, stirring occasionally, until the onions are soft and golden, about 10 minutes. Stir in the curry powder, tomato paste, and sea salt and continue cooking for 2-3 minutes, until fragrant. Next add 1/2 cup of water (adjust amount if altering batch size) to deglaze the pan. Stir and cook for about 3 more minutes until the water has reduced slightly.",
				"Place the onion and spice mixture into a high-speed blender with 2 cups of water (adjust amount if altering batch size) and the cashews. Blend on high until the mixture is smooth and creamy, about 2 minutes.",
				"Pour the contents of the blender back into the pan and bring to a simmer over medium heat. Reduce heat and simmer for 10-15 minutes, or until the sauce has thickened.",
				"COMBINE: Once the roasted cauliflower and tofu look crispy, add them to the sauce and reduce for about 5 minutes until the consistency is to your liking (we prefer it to be almost a cheese sauce consistency). Taste and adjust as needed, adding more salt for overall flavor or optional lemon juice for brightness.",
				"We like to serve it over rice with dairy-free yogurt, pickled onions, and cilantro (all optional, but extra delicious!). Store in a sealed container in the refrigerator for up to 4-5 days or in the freezer for 1 month (or longer).",
			},
		},
	},
	"cooking-nytimes": {
		Skip:      false,
		Callbacks: 1,
		Recipe: Recipe{
			SourceURL: mockURL,
			Title:     "Skillet Chicken With Mushrooms and Caramelized Onions",
			Ingredients: []string{
				"5 tablespoons olive oil", "2 tablespoons plus 2 teaspoons sherry vinegar",
				"2 teaspoons honey", "1 teaspoon Dijon mustard",
				"1/4 teaspoon red-pepper flakes", "Kosher salt and black pepper",
				"1 1/2 pounds boneless, skinless chicken thighs, cut into 3-inch pieces",
				"2 medium yellow onions, thinly sliced (about 4 cups)",
				"3/4 pound cremini mushrooms, stems removed and thinly sliced (about 4 cups)",
				"1/2 cup fresh flat-leaf parsley or dill leaves and fine stems, roughly chopped",
				"1/4 cup grated Parmesan or pecorino (optional)",
				"Bread or cooked pasta, for serving",
			},
			Instructions: []string{
				"In a large mixing bowl, combine 2 tablespoons oil, 2 tablespoons vinegar, the honey, mustard, red-pepper flakes and 1 teaspoon salt; whisk until smooth. Pat the chicken dry and season with salt and pepper, then add to the mixture, coating it well. Set aside at room temperature, stirring it once while you make the onions.",
				"Heat a 12-inch cast-iron or heavy skillet over medium-high until very hot, 1 1/2 to 2 minutes, then add the onions in an even layer. Season with salt, then cook, mostly undisturbed, for 4 minutes more, stirring every minute or so. Add the mushrooms, season with salt, and stir to combine. (It will look crowded, and that’s OK.) Allow to cook mostly undisturbed until the mushrooms shrink and start to brown, about 4 minutes, stirring every minute or so.",
				"Stir in the remaining 3 tablespoons olive oil and allow the onions to cook until they start to color, stirring and lowering the heat as necessary to avoid burning, about 2 minutes. Push the onions and mushrooms to the edges of the skillet, then add the chicken pieces to the center. Pour any remaining marinade (there will be very little) over the onions and mushrooms. Cook undisturbed for 4 to 5 minutes, then combine the chicken and vegetables and cook, stirring occasionally, until the chicken is cooked through, about 10 minutes more. (Reduce the heat to medium if the onions look like they are burning at any point.)",
				"Add the remaining 2 teaspoons sherry vinegar, stirring and scraping up anything on the bottom of the skillet. Season to taste with salt.",
				"Remove from the heat and top with the parsley and cheese, if using. Serve with bread or pasta.",
			},
		},
	},
	"ricardocuisine": {
		Skip:      false,
		Callbacks: 1,
		Recipe: Recipe{
			SourceURL: mockURL,
			Title:     "Rôti de porc braisé aux pommes",
			Ingredients: []string{
				"2,5 kg  (5 1/2 lb) d’épaule de porc avec la couenne et l’os  (ou échine)",
				"60 ml  (1/4 tasse) d’huile d’olive",
				"3   oignons, hachés",
				"3   gousses d’ail, hachées",
				"500 ml  (2 tasses) de bouillon de poulet",
				"375 ml  (1 1/2 tasse) de compote de pommes non sucrée",
				"15 ml  (1 c. à soupe) de moutarde à l’ancienne",
				"1 litre  (4 tasses) de pommes de terre pelées et coupées  en gros cubes ou de petites pommes de terre grelots entières",
				"60 ml  (1/4 tasse) de persil plat ciselé",
				"60 ml  (1/4 tasse) de ciboulette fraîche ciselée",
				"3   pommes Lobo ou Cortland épépinées et  coupées en 12 quartiers chacune",
				"15 ml  (1 c. à soupe) de miel",
				"Sel et poivre",
			},
			Instructions: []string{
				"Placer la grille au centre du four. Préchauffer le four \r\nà 180\u00a0°C (350\u00a0°F).",
				"Dans une grande casserole allant au four, dorer le r\u00f4ti sur toutes les faces dans la moiti\u00e9 de l\u2019huile (30 ml/2 c. \u00e0 soupe). Saler et poivrer. R\u00e9server la viande sur une assiette. Dans la m\u00eame casserole, dorer les oignons et l\u2019ail. Ajouter le bouillon, \r\nla compote de pommes et la moutarde. Remettre le r\u00f4ti dans la casserole c\u00f4t\u00e9 couenne vers le haut et porter \u00e0 \u00e9bullition.",
				"Couvrir et cuire au four 1\u00a0h\u00a030. Ajouter les pommes de \r\nterre autour de la viande et poursuivre la cuisson \u00e0 couvert 1\u00a0heure. Retirer le couvercle et poursuivre la cuisson 1\u00a0heure ou jusqu\u2019\u00e0 ce que la viande soit tendre et se d\u00e9fasse \u00e0 la fourchette. Ajouter les herbes.",
				"Entre-temps, dans une po\u00eale antiadh\u00e9sive, dorer les pommes dans le reste de l\u2019huile (30 ml/2 c. \u00e0 soupe) et le miel. Au moment de servir, placer les pommes autour du r\u00f4ti.",
			},
		},
	},
	"traeger": {
		Skip:      false,
		Callbacks: 3,
		Recipe: Recipe{
			SourceURL: mockURL,
			Title:     "BBQ Pulled Pork Sliders Recipe | Traeger Grills",
			Ingredients: []string{
				"Persian cucumbers, washed", "kosher salt", "Sugar", "garlic cloves",
				"dried bay leaves", "whole black peppercorns", "whole white peppercorns",
				"whole fennel seeds", "whole coriander seeds", "red pepper flakes",
				"white wine vinegar", "ice water", "fresh dill", "(8-10 lb) bone-in pork butt",
				"Traeger Pork & Poultry Rub, plus more to taste",
				"Traeger Apricot BBQ Sauce",
				"Mayonnaise",
				"ketchup",
				"Traeger Apricot BBQ Sauce",
				"sweet pickle relish",
				"yellow mustard",
				"Sugar",
				"distilled white vinegar",
				"Slider Buns",
			},
			Instructions: []string{
				"When ready to cook, set Traeger temperature 165℉ and preheat, lid closed for 15 minutes. For optimal flavor, use Super Smoke if available.",
				"For the Homemade Pickles: Plan ahead, these pickles brine overnight. Place the whole, washed cucumbers directly on the grill grate and smoke for 30 to 45 minutes.",
				"In a medium saucepan, combine the salt, sugar, garlic, bay leaves, black and white peppercorns, fennel seeds, coriander seeds, red pepper flakes, and white wine vinegar. Bring to a boil over medium-high heat, then reduce the heat to medium-low and simmer for 5-10 minutes, stirring to dissolve the salt and sugar. Remove the brine from the heat and let steep for 20 minutes more. Pour the ice water into the brine and add the dill.",
				"Remove the cucumbers from the Traeger and thinly slice crosswise. Transfer to a large glass jar or non-reactive container.",
				"Pour the brine over the cucumbers and weigh down as needed to ensure the cucumbers are fully submerged. Cover and let sit at room temperature for 24 hours before using.",
				"Make the pulled pork: Generously the pork butt on all sides with Traeger Pork & Poultry Rub. Let sit at room temperature for 45 minutes.",
				"When ready to cook, set the Traeger temperature to 225℉ and preheat with the lid closed for 15 minutes. For optimal flavor, use Super Smoke, if available.",
				"Place the pork butt directly on the grill grates, fat-side down. Close the lid and smoke for 3 hours.",
				"Transfer the pork to a disposable aluminum foil pan and cover the pan with foil. Increase the Traeger temperature to 275℉. Place the pan on the grill, close the lid, and roast until an instant-read thermometer inserted in the thickest part of the roast, but not touching bone, registers 200℉, 3-5 hours more.",
				"Meanwhile, make the special sauce: In a small bowl, stir together the mayonnaise, ketchup, Traeger Apricot BBQ Sauce, sweet pickle relish, mustard, sugar, and distilled white vinegar. Refrigerate until ready to use.",
				"Carefully transfer the pork roast to a clean disposable pan and let rest for 20 minutes. Pour the juices from the bottom of the roasting pan into a fat separator.",
				"With your hands (preferably protected from the heat with lined, heavy-duty rubber gloves), pull the pork into chunks. Discard the bone and any lumps of fat, including the cap. Pull each chunk into shreds and transfer to a large bowl. Season with Traeger Pork & Poultry Rub to taste and moisten with the reserved pork juices, discard any fat that has floated to the top. Add the Traeger Apricot BBQ Sauce and mix well.",
				"Pile the pulled pork on the slider buns. Top with the pickles and special sauce. Enjoy!",
			},
		},
	},
	"4ks": {
		Skip:      false,
		Callbacks: 1,
		Recipe: Recipe{
			SourceURL: mockURL,
			Title:     "Overnight chocolate oats",
			Ingredients: []string{
				"1 container Silken tofu",
				"1 cup Steal cut oats",
				"2 tbsp Coco powder",
				"2 tbsp Maple syrup",
				" ",
				"1 Banana",
				"1 tbsp Cinamon",
				"2 tbsp Chia seeds",
				"2 tbsp Flax seed",
				" Peanut butter",
			},
			Instructions: []string{
				"Put everything in the blinder except peanut butter.",
				"Blend like it's nobody's business.",
				"Put in container.",
				"Drizzle your peanut butter on it.",
				"Make some design lines with a nife (it's to mix the PB lightly).",
				"Leave it until tomorrow.",
			},
		},
	},
	"allrecipes": {
		Skip:      false,
		Callbacks: 1,
		Recipe: Recipe{
			SourceURL: mockURL,
			Title:     "Good Old-Fashioned Pancakes",
			Ingredients: []string{
				"1.5 cups all-purpose flour",
				"3.5 teaspoons baking powder",
				"1 tablespoon white sugar",
				"0.25 teaspoon salt, or more to taste",
				"1.25 cups milk",
				"3 tablespoons butter, melted",
				"1 egg",
			},
			Instructions: []string{
				"Sift flour, baking powder, sugar, and salt together in a large bowl. Make a well in the center and add milk, melted butter, and egg; mix until smooth.",
				"Heat a lightly oiled griddle or pan over medium-high heat. Pour or scoop the batter onto the griddle, using approximately 1/4 cup for each pancake; cook until bubbles form and the edges are dry, about 2 to 3 minutes. Flip and cook until browned on the other side. Repeat with remaining batter.",
			},
		},
	},
	"recettes.qc.ca": {
		Skip:      true,
		Callbacks: 1,
		Recipe: Recipe{
			SourceURL:    mockURL,
			Title:        "Overnight chocolate oats",
			Ingredients:  []string{},
			Instructions: []string{},
		},
	},
}

func TestRecipeHTML(t *testing.T) {
	// run through test cases
	for name, tc := range testCases {
		// sub tests
		t.Run(name, func(t *testing.T) {
			if tc.Skip {
				t.Skipf("Skipping test %s", name)
			}
			callbackCount := 0

			// read test html
			dat, err := os.ReadFile(fmt.Sprintf("./testdata/%s.html", name))
			if err != nil {
				t.Errorf("(%s) failed to read file: %v", name, err)
			}

			// start mock html server
			ts := newTestHTTPServer(dat)
			defer ts.Close()
			u := ts.URL + "/recipe"

			// create colly collector
			c := colly.NewCollector()

			// onHTML callback
			c.OnHTML(ApplicationJSONLDScriptTag, func(e *colly.HTMLElement) {
				callbackCount++

				r, err := getRecipeFromJSONLD(e, u)
				if err == errNotRecipe {
					return
				}
				if err != nil {
					t.Errorf("(%s) failed to get recipe from jsonld: %v", name, err)
				}
				diff := cmp.Diff(r, tc.Recipe)
				if diff != "" {
					t.Errorf("(%s) unexpected number ld+json scripts detected. %v", name, diff)
				}
				// PrintStruct(r)
			})

			// visit
			c.Visit(u)

			// validate
			diff := cmp.Diff(callbackCount, tc.Callbacks)
			if diff != "" {
				t.Errorf("(%s) unexpected number ld+json scripts detected. %v", name, diff)
			}

		})
	}

}
