function showGame() {
    const question = document.getElementsByClassName("question")[0];
    question.style.visibility = 'visible';
    const answer = document.getElementsByClassName("answers")[0];
    answer.style.display = 'flex';
}

function hideGame() {
    const question = document.getElementsByClassName("question")[0];
    question.style.visibility = 'hidden';
    const answer = document.getElementsByClassName("answers")[0];
    answer.style.display = 'none';
}

function startTimer(maxSteps, stepTime, onTimeout) {
    const timerBar = document.getElementById('timer-bar');
    timerBar.style.width = '100%';
    let i = 0;
    const timer = setInterval(() => {
        i++;
        timerBar.style.width = Math.round((1 - i / maxSteps) * 100) + '%';
        if (i === maxSteps) {
            clearInterval(timer);
            onTimeout();
        }
    }, stepTime);
    return timer;
}

function clearAnswerOnClicks() {
    const aEls = document.getElementById("answers").children;
    for (const aEl of aEls) {
        aEl.onclick = null;
    }
}

function clearAnswers() {
    const ansEl = document.getElementById("answers");
    while (ansEl.children.length > 0) {
        ansEl.removeChild(ansEl.children[0]);
    }
    document.getElementById('timer-bar').style.width = '100%';
}

function showRightAnswer(index) {
    const aEls = document.getElementById("answers").children;
    aEls[index].classList.add('right-answer');
}

function setRandomQuestion(ctx) {
    const {questions} = ctx.topic;
    const ind = Math.floor(Math.random() * questions.length);
    const question = questions[ind];
    const rightInd = Math.floor(Math.random() * question.answersNumber);
    const answers = new Array(question.answersNumber);
    let wrongAnswersLeft = question.wrongAnswers.length;
    for (let i = 0; i < question.answersNumber; i++) {
        if (i === rightInd) {
            answers[i] = question.rightAnswer;
        } else {
            const ind = Math.floor(Math.random() * wrongAnswersLeft);
            answers[i] = question.wrongAnswers[ind];
            const tmp = question.wrongAnswers[ind];
            question.wrongAnswers[ind] = question.wrongAnswers[wrongAnswersLeft - 1];
            question.wrongAnswers[wrongAnswersLeft - 1] = tmp;
            wrongAnswersLeft--;
        }
    }
    const timer = startTimer(3 * 60, 33, () => {
        clearAnswerOnClicks();
        showRightAnswer(rightInd);
        setTimeout(showStartMenu, 3000);
    });
    setQuestion(ctx, question.text, answers, rightInd, (good) => {
        clearInterval(timer);
        if (!good) {
            setTimeout(showStartMenu, 3000);
        }
    });
}

function setQuestion(ctx, question, answers, rightAnswerInd, saveResult) {
    const qEl = document.getElementById("question");
    const ansEl = document.getElementById("answers");
    qEl.innerText = question;
    const aEls = new Array(answers.length);
    for (let i = 0; i < aEls.length; i++) {
        aEls[i] = document.createElement('div');
        aEls[i].innerText = answers[i];
        aEls[i].onclick = function () {
            aEls[rightAnswerInd].classList.add('right-answer');
            if (i === rightAnswerInd) {
                setTimeout(() => {
                    for (const aEl of aEls) {
                        aEl.classList.remove('right-answer', 'wrong-answer');
                    }
                    setRandomQuestion(ctx);
                }, 1000);
                ctx.score += 5;
                document.getElementById('score').innerText = '' + ctx.score;
                const u = new URL(window.location.href);
                const body = {
                    userId: +u.searchParams.get('userId'),
                    score: +ctx.score,
                };
                if (u.searchParams.get('inlineId') !== null) {
                    body['inlineId'] = u.searchParams.get('inlineId');
                }
                if (u.searchParams.get('chatId') !== null) {
                    body['chatId'] = u.searchParams.get('chatId');
                }
                if (u.searchParams.get('messageId') !== null) {
                    body['messageId'] = u.searchParams.get('messageId');
                }
                fetch('/api/', {
                    method: 'POST',
                    body: JSON.stringify(body),
                }).catch((...args) => {
                    console.log('Result: ', ...args);
                });
            } else {
                aEls[i].classList.add('wrong-answer');
            }
            saveResult(i === rightAnswerInd);
            clearAnswerOnClicks();
        };
    }
    clearAnswers();
    for (let i = 0; i < aEls.length; i++) {
        ansEl.appendChild(aEls[i]);
    }
}

function showStartMenu() {
    clearAnswers();
    hideGame();
    const startButton = document.getElementById("start-button");
    startButton.style.display = 'block';
}

window.addEventListener('load', () => {
    const startButton = document.getElementById("start-button");

    const u = new URL(window.location.href);
    const topicId = u.searchParams.get('topicId');
    fetch(`/api/topics/${topicId}`).then(resp => resp.json()).then(topic => {
        startButton.addEventListener('click', () => {
            let ctx = {score: 0, topic};
            startButton.style.display = 'none';
            document.getElementById("score").innerText = '0';
            setRandomQuestion(ctx);
            showGame();
        });
        showStartMenu();
    }).catch(console.log);
});
