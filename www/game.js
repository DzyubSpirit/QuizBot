function changeGameVisibility(visibility) {
    const question = document.getElementsByClassName("question")[0];
    const answer = document.getElementsByClassName("answers")[0];
    const els = [question, answer];
    for (const el of els) {
        el.style.visibility = visibility;
    }
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

function clearAnswersHighlighting() {
    const aEls = document.getElementById("answers").children;
    for (const aEl of aEls) {
        aEl.classList.remove('right-answer', 'wrong-answer');
    }
    document.getElementById('timer-bar').style.width = '100%';
}

function setRandomQuestion(ctx) {
    const books = ["Бытие", "Исход", "Левит", "Числа", "Второзаконие", "Иисус Навин", "Судьей", "Руфь", "1 Царств", "2 Царств", "3 Царств", "4 Царств", "1 Паралипоменон", "2 Паралипоменон", "Ездра", "Неемия", "Есфирь", "Иов", "Псалтырь", "Притчи", "Екклесиаст", "Книга Песнь Песней", "Исаия", "Иеремия", "Плач Иеремии", "Иезекииль", "Даниил", "Осия", "Иоиль", "Амос", "Авдий", "Иона", "Михей", "Наум", "Аввакум", "Софония", "Аггей", "Захария", "Малахия", "Матфея", "Марка", "Луки", "Иоанна", "Деяния", "Иакова", "1 Петра", "2 Петра", "1 Иоанна", "2 Иоанна", "3 Иоанна", "Иуды", "Римлянам", "1 Коринфянам", "2 Коринфянам", "Галатам", "Ефесянам", "Филиппийцам", "Колоссянам", "1 Фессалоникийцам", "2 Фессалоникийцам", "1 Тимофею", "2 Тимофею", "Титу", "Филимону", "Евреям", "Откровение"];
    // const books = ["Книга Бытие", "Книга Исход", "Книга Левит", "Книга Числа", "Книга Второзаконие", "Книга Иисуса Навина", "Книга судей израилевых", "Книга Руфь", "Первая книга царств", "Вторая книга царств", "Третья книга царств", "Четвертая книга царств", "Первая книга Паралипоменон", "Вторая книга Паралипоменон", "Книга Ездры", "Книга Неемии", "Книга Есфирь", "Книга Иова", "Псалтирь", "Книга притчей соломоновых", "Книга Екклесиаста", "Книга песни песней Соломона", "Книга пророка Исаии", "Книга пророка Иеремии", "Книга плач Иеремии", "Книга пророка Иезекииля", "Книга пророка Даниила", "Книга пророка Осии", "Книга пророка Иоиля", "Книга пророка Амоса", "Книга пророка Авдия", "Книга пророка Ионы", "Книга пророка Михея", "Книга пророка Наума", "Книга пророка Аввакума", "Книга пророка Софонии", "Книга пророка Аггея", "Книга пророка Захарии", "Книга пророка Малахии", "Евангелие от Матфея", "Евангелие от Марка", "Евангелие от Луки", "Евангелие от Иоанна", "Деяния апостолов", "Послание Иакова", "Первое послание Петра", "Второе послание Петра", "Первое послание Иоанна", "Второе послание Иоанна", "Третье послание Иоанна", "Послание Иуды", "Послание к римлянам", "Первое послание к коринфянам", "Второе послание к коринфянам", "Послание к галатам", "Послание к ефесянам", "Послание к филиппийцам", "Послание к колоссянам", "Первое послание к фессалоникийцам", "Второе послание к фессалоникийцам", "Первое послание к Тимофею", "Второе послание  к Тимофею", "Послание к Титу", "Послание к Филимону", "Послание к евреям", "Откровение Иоанна"];
    const ind = Math.floor(Math.random() * books.length);
    const rightInd = (ind + 1) % books.length;
    const answers = new Array(5);
    for (let i = 0; i < 5; i++) {
        let rand = Math.floor(Math.random() * books.length);
        while (rand === rightInd) {
            rand = Math.floor(Math.random() * books.length);
        }
        answers[i] = books[rand];
    }
    const rightAnswerInd = Math.floor(Math.random() * 5);
    answers[rightAnswerInd] = books[rightInd];
    const timer = startTimer(3 * 60, 33, () => {
        clearAnswerOnClicks();
        setTimeout(showStartMenu, 3000);
    });
    setQuestion(ctx, books[ind], answers, rightAnswerInd, (good) => {
        clearInterval(timer);
        if (!good) {
            setTimeout(showStartMenu, 3000);
        }
    });
}

function setQuestion(ctx, question, answers, rightAnswerInd, saveResult) {
    const qEl = document.getElementById("question");
    qEl.innerText = `Какая книга идет после \n"${question}"`;
    const aEls = document.getElementById("answers").children;
    for (let i = 0; i < aEls.length; i++) {
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
}

function showStartMenu() {
    clearAnswersHighlighting();
    changeGameVisibility('hidden');
    const startButton = document.getElementById("start-button");
    startButton.style.visibility = 'visible';
}

window.addEventListener('load', () => {
    const startButton = document.getElementById("start-button");
    startButton.addEventListener('click', () => {
        let ctx = {score: 0};
        startButton.style.visibility = 'hidden';
        document.getElementById("score").innerText = '0';
        setRandomQuestion(ctx);
        changeGameVisibility('visible');
    });
    showStartMenu();
});
