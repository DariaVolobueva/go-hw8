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

    _, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Канал для сигналізації про завершення
    done := make(chan struct{})

    // Обробка сигналу переривання
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt)
        <-sigChan
        fmt.Println("\nОтримано сигнал переривання. Завершення програми...")
        cancel()
        close(done)
    }()

    roundNum := 1
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    fmt.Println("Гра починається!")

    for {
        select {
        case <-done:
            fmt.Println("Гру завершено.")
            return
        case <-ticker.C:
            question := generateRound(roundNum, rng)
            fmt.Printf("\nРаунд %d: %s\n", roundNum, question.Text)

            answerCounts := make([]int, 4)
            playerResults := make(map[int]bool)

            var wg sync.WaitGroup
            for _, player := range players {
                wg.Add(1)
                go func(p Player) {
                    defer wg.Done()
                    answer := playerAnswer(rng)
                    answerCounts[answer]++
                    playerResults[p.ID] = (answer == question.Correct)
                }(player)
            }
            wg.Wait()

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