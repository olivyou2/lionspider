# LionSpider
인터넷을 탐색하면서, 사이트 정보를 수집합니다. 

#### 주의사항
robots.txt 파일의 크롤링 규칙을 잘 지켜줍시다
[에를들어 이런 곳!](https://gbsb.tistory.com/80)

## How to Start
````bash
git clone https://github.com/olivyou2/lionspider
cd lionsider
go run .
````

## How LionSpider Works
### 1. Crawl Site
LionSpider 에 진입점이 주어지고 작업을 시작하면, 진입점을 Html Parser 를 통해 파싱하여 Innertext 를 모두 구해옵니다. Innertext 에서 필요없는 문자 ( \r, \n, \t ... ) 을 제거하고, 공백으로 구분하여 사이트에 단어를 매칭시킵니다. 여기서 단어는 Embeding 되어 Vector 로 매칭하게 됩니다.

### 2. Explore Other Sites
아까 얻어낸 사이트에서, 정규표현식 "http([^"]\*)" 을 통해 링크를 크롤링합니다. 크롤링된 링크를 LinkQueue 의 끝에 넣습니다.

### 3. Choose Site
이전 크롤링 작업이 끝나고, 새로운 크롤링 작업을 시작할 때, CrawlPolicy 에 따라 LinkQueue 내부에서 중요도가 높은 링크를 선별합니다. 현재 링크선별 과정에서는, 초기에 설정했던 단어가 링크에 몇 개 들어있는지로 중요도를 판별하고, Sort 작업을 진행합니다.

### 4. Recursion
그렇게 선별된 링크를 다시 크롤링 합니다

## Multicore Processing
멀티코어.. 까지는 아니고 goroutine 을 통해 병렬적으로 크롤링 작업을 수행할 수 있습니다. 메모리는 mutex 를 통해 관리됩니다.
