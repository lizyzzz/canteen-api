INSERT INTO `dishes` (`name`, `price`, `category`, `ingredients`, `image_url`)
VALUES
    ('红排排骨', 38.00, '热销', 'xxx', '/images/dish1.jpg'),
    ('宫保鸡丁', 32.00, '热销', 'xxx', '/images/dish2.jpg'),
    ('炒米粉', 9.99, '主食', 'xxx', '/images/dish3.jpg'),
    ('薯条', 10.00, '小吃', 'Beef Patty, Lettuce, Tomato, Cheese, Bun', '/images/dish4.jpg'),
    ('可乐', 6.00, '饮品', 'Sushi Rice, Raw Fish, Seaweed, Wasabi, Soy Sauce', '/images/dish5.jpg'),
    ('鸡腿套餐', 28.88, '套餐', 'Sushi Rice, Raw Fish, Seaweed, Wasabi, Soy Sauce', '/images/dish6.jpg');


select o.id, o.user_id, o.status, o.create_time, o.pickup_time, o.total_price, 
item.dish_id, item.dish_name, item.unit_price, item.quantity, item.subtotal 
from orders o 
left join order_item item 
on o.id = item.order_id
where o.id in (1, 2, 3)

update user set balance = IF(balance + -1000 < 0, 0, balance + -1000)
where id = 4