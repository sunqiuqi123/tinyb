/* 计算 1+2+...+10 */
main() {
    auto sum, i;
    sum = 0;
    i = 1;
    while (i <= 10) {
        sum = sum + i;
        i = i + 1;
    }
    putnumb(sum);
    putchar('\n');
}
