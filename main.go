package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Player struct {
	ID int
}

type Question struct {
	Text     string
	Options  []string
	Correct  int
	RoundNum int
}

type RoundResult struct {
	RoundNum      int
	AnswerCounts  []int
	PlayerResults map[int]bool
}

func generateRound(roundNum int, rng *rand.Rand) Question {
	return Question{
		Text:     fmt.Sprintf("Питання %d", roundNum),
		Options:  []string{"A", "B", "C", "D"},
		Correct:  rng.Intn(4),
		RoundNum: roundNum,
	}
}

func playerAnswer(rng *rand.Rand) int {
	return rng.Intn(4)
}

func main() {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	playerCount := 5
	players := make([]Player, playerCount)
	for i := range players {
		players[i] = Player{ID: i}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan
		fmt.Println("\nОтримано сигнал переривання. Завершення програми...")
		cancel()
		close(done)
	}()

	type playerInput struct {
		question Question
		resultCh chan<- int
	}
	playerChannels := make([]chan playerInput, playerCount)
	for i := range playerChannels {
		playerChannels[i] = make(chan playerInput)
	}

	var wg sync.WaitGroup
	for i, player := range players {
		wg.Add(1)
		go func(p Player, ch <-chan playerInput) {
			defer wg.Done()
			playerRng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for {
				select {
				case <-ctx.Done():
					return
				case input := <-ch:
					answer := playerAnswer(playerRng)
					input.resultCh <- answer
				}
			}
		}(player, playerChannels[i])
	}

	roundNum := 1
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	fmt.Println("Гра починається!")

	for {
		select {
		case <-done:
			fmt.Println("Гру завершено.")
			cancel()
			wg.Wait()
			return
		case <-ticker.C:
			question := generateRound(roundNum, rng)
			fmt.Printf("\nРаунд %d: %s\n", roundNum, question.Text)

			answerCounts := make([]int, 4)
			playerResults := make(map[int]bool)

			resultCh := make(chan int, playerCount)
			for i := range players {
				playerChannels[i] <- playerInput{question: question, resultCh: resultCh}
			}

			for i := 0; i < playerCount; i++ {
				answer := <-resultCh
				answerCounts[answer]++
				playerResults[i] = (answer == question.Correct)
			}

			fmt.Println("Результати раунду:")
			for i, count := range answerCounts {
				fmt.Printf("Варіант %s: %d відповідей\n", fmt.Sprintf("%c", 'A'+i), count)
			}
			fmt.Println("Результати гравців:")
			for playerID, correct := range playerResults {
				if correct {
					fmt.Printf("Гравець %d: Правильно\n", playerID)
				} else {
					fmt.Printf("Гравець %d: Неправильно\n", playerID)
				}
			}

			roundNum++
		}
	}
}