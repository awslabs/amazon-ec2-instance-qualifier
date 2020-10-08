package database

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
)

type ranges struct {
	startingID int //inclusive
	endingID   int //exclusive
}

// PopulateTable populates table in bulk
func PopulateTable(numEntries int) error {
	var wg sync.WaitGroup
	numCpus := runtime.NumCPU()

	chunkedWork := chunk(numEntries, numCpus)
	fmt.Printf("chunkedWork: %v\n", chunkedWork)
	if len(chunkedWork) < 1 {
		return fmt.Errorf("there was an error distributing work")
	}

	for i := 0; i < numCpus; i++ {
		wg.Add(1)
		go func(idRanges ranges) {
			defer wg.Done()
			for j := idRanges.startingID; j < idRanges.endingID; j++ {
				randomBreed := rand.Intn(len(dogBreeds))
				randomName := rand.Intn(len(names))
				randomStatus := rand.Intn(len(statuses))
				entry := Pet{
					PetId:  j,
					Name:   names[randomName],
					Breed:  dogBreeds[randomBreed],
					Status: statuses[randomStatus],
				}
				err := AddPet(entry)
				if err != nil {
					fmt.Println("There was an error adding to Pets table: ", err.Error())
				}
			}
		}(chunkedWork[i])
	}
	wg.Wait()
	return nil
}

// chunk distributes work based on number of cpus
func chunk(numEntries, numCpus int) (result []ranges) {
	chunkSize := (numEntries + numCpus - 1) / numCpus
	if chunkSize < 1 {
		chunkSize = 1
	}
	startId := 0
	endId := chunkSize

	for itemsRemaining := numEntries; itemsRemaining > 0; itemsRemaining -= chunkSize {
		if endId > numEntries {
			endId = numEntries
		}
		idRange := ranges{
			startingID: startId,
			endingID:   endId,
		}
		result = append(result, idRange)
		startId = endId
		endId += chunkSize
	}

	return result
}

var dogBreeds = []string{
	"Affenpinscher",
	"Afghan hound",
	"African hunting dog",
	"Airedale",
	"American Staffordshire terrier",
	"Appenzeller",
	"Australian terrier",
	"Basenji",
	"Basset",
	"Beagle",
	"Bedlington terrier",
	"Bernese mountain dog",
	"Black-and-tan coonhound",
	"Blenheim spaniel",
	"Bloodhound",
	"Bluetick",
	"Border collie",
	"Border terrier",
	"Borzoi",
	"Boston bull",
	"Bouvier des Flandres",
	"Boxer",
	"Brabancon griffon",
	"Briard",
	"Brittany spaniel",
	"Bull mastiff",
	"Cairn",
	"Cardigan",
	"Chesapeake Bay retriever",
	"Chihuahua",
	"Chow",
	"Clumber",
	"Cocker spaniel",
	"Collie",
	"Curly-coated retriever",
	"Dandie Dinmont",
	"Dhole",
	"Dingo",
	"Doberman",
	"English foxhound",
	"English setter",
	"English springer",
	"EntleBucher",
	"Eskimo dog",
	"Flat-coated retriever",
	"French bulldog",
	"German shepherd",
	"German short-haired pointer",
	"Giant schnauzer",
	"Golden retriever",
	"Gordon setter",
	"Great Dane",
	"Great Pyrenees",
	"Greater Swiss Mountain dog",
	"Groenendael",
	"Ibizan hound",
	"Irish setter",
	"Irish terrier",
	"Irish water spaniel",
	"Irish wolfhound",
	"Italian greyhound",
	"Japanese spaniel",
	"Keeshond",
	"Kelpie",
	"Kerry blue terrier",
	"Komondor",
	"Kuvasz",
	"Labrador retriever",
	"Lakeland terrier",
	"Leonberg",
	"Lhasa",
	"Malamute",
	"Malinois",
	"Maltese dog",
	"Mexican hairless",
	"Miniature pinscher",
	"Miniature poodle",
	"Miniature schnauzer",
	"Mut",
	"Newfoundland",
	"Norfolk terrier",
	"Norwegian elkhound",
	"Norwich terrier",
	"Old English sheepdog",
	"Otterhound",
	"Papillon",
	"Pekinese",
	"Pembroke",
	"Pomeranian",
	"Pug",
	"Redbone",
	"Rhodesian ridgeback",
	"Rottweiler",
	"Saint Bernard",
	"Saluki",
	"Samoyed",
	"Schipperke",
	"Scotch terrier",
	"Scottish deerhound",
	"Sealyham terrier",
	"Shetland sheepdog",
	"Shih-Tzu",
	"Siberian husky",
	"Silky terrier",
	"Soft-coated wheaten terrier",
	"Staffordshire bullterrier",
	"Standard poodle",
	"Standard schnauzer",
	"Sussex spaniel",
	"Tibetan mastiff",
	"Tibetan terrier",
	"Toy poodle",
	"Toy terrier",
	"Vizsla",
	"Walker hound",
	"Weimaraner",
	"Welsh springer spaniel",
	"West Highland white terrier",
	"Whippet",
	"Wire-haired fox terrier",
	"Yorkshire terrier",
}

var names = []string{
	"Adiana",
	"Adina",
	"Adora",
	"Adore",
	"Adoree",
	"Adorne",
	"Adrea",
	"Adria",
	"Belva",
	"Belvia",
	"Bendite",
	"Benetta",
	"Benita",
	"Benni",
	"Bennie",
	"Benny",
	"Benoite",
	"Berenice",
	"Carmon",
	"Caro",
	"Carol",
	"Carol-Jean",
	"Carola",
	"Carrie",
	"Carrissa",
	"Carroll",
	"Carry",
	"Cary",
	"Caryl",
	"Caryn",
	"Casandra",
	"Casey",
	"Casi",
	"Casie",
	"Cass",
	"Demetris",
	"Dena",
	"Deni",
	"Denice",
	"Denise",
	"Denna",
	"Denni",
	"Dennie",
	"Denny",
	"Elayne",
	"Elberta",
	"Eleanora",
	"Eleanore",
	"Electra",
	"Fredericka",
	"Frederique",
	"Fredi",
	"Fredia",
	"Fredra",
	"Fredrika",
	"Freida",
	"Gene",
	"Geneva",
	"Genevieve",
	"Genevra",
	"Genia",
	"Genna",
	"Genvieve",
	"Harriette",
	"Harriot",
	"Harriott",
	"Hatti",
	"Hatty",
	"Ida",
	"Idalia",
	"Idalina",
	"Idaline",
	"Idell",
	"Idelle",
	"Idette",
	"Ileana",
	"Ileane",
	"Ilene",
	"Jaclin",
	"Jaclyn",
	"Jacquelin",
	"Jacqueline",
	"Jacquelyn",
	"Jacquelynn",
	"Jacquenetta",
	"Jacquenette",
	"Jacquetta",
	"Jacquette",
	"Jacqui",
	"Jacquie",
	"Jacynth",
	"Jada",
	"Karalee",
	"Karalynn",
	"Kare",
	"Karee",
	"Karel",
	"Karen",
	"Karena",
	"Kari",
	"Karia",
	"Tatum",
	"Tawnya",
	"Tawsha",
	"Ted",
	"Tedda",
	"Teddi",
	"Teddie",
	"Teddy",
	"Tedi",
	"Tedra",
	"Wendy",
	"Wendye",
	"Yalonda",
	"Yasmeen",
	"Yasmin",
	"Yelena",
	"Yetta",
	"Yettie",
	"Yetty",
	"Zabrina",
}

var statuses = []string{
	"Available",
	"Pending",
	"Sold",
}
